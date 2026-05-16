package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"socket-flow/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

func TestKeycloakAuthenticatorValidate(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}

	const kid = "test-key"
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(jwksDocument{
			Keys: []jwk{{
				Kid: kid,
				Kty: "RSA",
				Alg: "RS256",
				Use: "sig",
				N:   base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes()),
				E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes()),
			}},
		}); err != nil {
			t.Fatalf("encode jwks: %v", err)
		}
	}))
	defer jwksServer.Close()

	cfg := config.KeycloakConfig{
		Issuer:            "https://keycloak.example/realms/socket-flow",
		JWKSURL:           jwksServer.URL,
		ClientID:          "socket-flow-api",
		Audience:          "socket-flow-api",
		AllowedAlgorithms: "RS256",
		JWKSCacheTTL:      "10m",
		ClockSkew:         "30s",
	}

	claims := KeycloakClaims{
		PreferredUsername: "farukh",
		Email:             "farukh@example.com",
		RealmAccess:       keycloakRoleClaim{Roles: []string{"user"}},
		ResourceAccess: map[string]keycloakRoleClaim{
			"socket-flow-api": {Roles: []string{"admin"}},
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "keycloak-subject",
			Issuer:    cfg.Issuer,
			Audience:  jwt.ClaimStrings{cfg.Audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("SignedString: %v", err)
	}

	user, err := NewKeycloakAuthenticator(cfg).Validate(context.Background(), tokenString)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}

	if user.Subject != "keycloak-subject" {
		t.Fatalf("Subject = %q; want keycloak-subject", user.Subject)
	}
	if user.Username != "farukh" || user.Email != "farukh@example.com" {
		t.Fatalf("user claims = %#v", user)
	}
	if len(user.Roles) != 2 {
		t.Fatalf("roles = %#v; want realm and client roles", user.Roles)
	}
}
