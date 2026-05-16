package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"socket-flow/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

const (
	defaultHTTPTimeout = 5 * time.Second
)

type KeycloakAuthenticator struct {
	cfg        config.KeycloakConfig
	httpClient *http.Client

	mu        sync.RWMutex
	keys      map[string]any
	expiresAt time.Time
}

type KeycloakClaims struct {
	PreferredUsername string                       `json:"preferred_username"`
	Email             string                       `json:"email"`
	RealmAccess       keycloakRoleClaim            `json:"realm_access"`
	ResourceAccess    map[string]keycloakRoleClaim `json:"resource_access"`
	jwt.RegisteredClaims
}

type keycloakRoleClaim struct {
	Roles []string `json:"roles"`
}

type jwksDocument struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	Kid string   `json:"kid"`
	Kty string   `json:"kty"`
	Alg string   `json:"alg"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5C []string `json:"x5c"`
}

func NewKeycloakAuthenticator(cfg config.KeycloakConfig) *KeycloakAuthenticator {
	return &KeycloakAuthenticator{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
		keys: make(map[string]any),
	}
}

func (a *KeycloakAuthenticator) Validate(ctx context.Context, tokenString string) (AuthenticatedUser, error) {
	claims := new(KeycloakClaims)

	parserOptions := []jwt.ParserOption{
		jwt.WithIssuer(a.cfg.Issuer),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(a.cfg.ClockSkewDuration()),
	}
	if a.cfg.Audience != "" {
		parserOptions = append(parserOptions, jwt.WithAudience(a.cfg.Audience))
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, a.keyfunc(ctx), parserOptions...)
	if err != nil {
		return AuthenticatedUser{}, errors.Wrap(err, "validate keycloak token")
	}
	if !token.Valid || claims.Subject == "" {
		return AuthenticatedUser{}, errors.New("invalid keycloak token claims")
	}

	return AuthenticatedUser{
		Subject:  claims.Subject,
		Username: claims.PreferredUsername,
		Email:    claims.Email,
		Roles:    claims.rolesForClient(a.cfg.ClientID),
	}, nil
}

func (a *KeycloakAuthenticator) keyfunc(ctx context.Context) jwt.Keyfunc {
	return func(token *jwt.Token) (any, error) {
		if !a.cfg.AllowsAlgorithm(token.Method.Alg()) {
			return nil, errors.Errorf("unexpected signing algorithm %q", token.Method.Alg())
		}

		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, errors.New("token header missing kid")
		}

		key, ok := a.cachedKey(kid)
		if ok {
			return key, nil
		}

		if err := a.refreshKeys(ctx); err != nil {
			return nil, err
		}

		key, ok = a.cachedKey(kid)
		if !ok {
			return nil, errors.Errorf("jwks key %q not found", kid)
		}

		return key, nil
	}
}

func (a *KeycloakAuthenticator) cachedKey(kid string) (any, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if time.Now().After(a.expiresAt) {
		return nil, false
	}

	key, ok := a.keys[kid]
	return key, ok
}

func (a *KeycloakAuthenticator) refreshKeys(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.cfg.JWKSURL, nil)
	if err != nil {
		return errors.Wrap(err, "build jwks request")
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "fetch jwks")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("jwks endpoint returned status %d", resp.StatusCode)
	}

	var doc jwksDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return errors.Wrap(err, "decode jwks")
	}

	keys := make(map[string]any, len(doc.Keys))
	for _, rawKey := range doc.Keys {
		if rawKey.Kid == "" || rawKey.Kty != "RSA" {
			continue
		}

		publicKey, err := rawKey.publicKey()
		if err != nil {
			return errors.Wrapf(err, "parse jwks key %q", rawKey.Kid)
		}

		keys[rawKey.Kid] = publicKey
	}

	a.keys = keys
	a.expiresAt = time.Now().Add(a.cfg.CacheDuration())

	return nil
}

func (k jwk) publicKey() (*rsa.PublicKey, error) {
	if len(k.X5C) > 0 {
		rawCert, err := base64.StdEncoding.DecodeString(k.X5C[0])
		if err != nil {
			return nil, errors.Wrap(err, "decode x5c certificate")
		}

		cert, err := x509.ParseCertificate(rawCert)
		if err != nil {
			return nil, errors.Wrap(err, "parse x5c certificate")
		}

		publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("x5c certificate is not RSA")
		}

		return publicKey, nil
	}

	modulusBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, errors.Wrap(err, "decode modulus")
	}

	exponentBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, errors.Wrap(err, "decode exponent")
	}

	exponent := new(big.Int).SetBytes(exponentBytes)
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(modulusBytes),
		E: int(exponent.Int64()),
	}, nil
}

func (c KeycloakClaims) rolesForClient(clientID string) []string {
	roles := append([]string{}, c.RealmAccess.Roles...)

	if clientID != "" {
		if access, ok := c.ResourceAccess[clientID]; ok {
			roles = append(roles, access.Roles...)
		}
	}

	return roles
}

func BearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}
