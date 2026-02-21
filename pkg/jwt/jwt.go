package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"math"
	"time"

	"github.com/caitlinelfring/go-env-default"
	"github.com/golang-jwt/jwt/v5"
	
	errs "socket-flow/internal/errors"
)

var (
	secret  = env.GetDefault("SECRET", "")
	baseUrl = env.GetDefault("BASE_URL", "http://localhost:8080")
)

type Claims struct {
	jwt.RegisteredClaims
	Role string
}

func generateTokenID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func GenerateJWT(subject string, minutes uint, role string) (string, error) {

	tokenID, err := generateTokenID()
	if err != nil {
		return "", err
	}

	issuedAt := time.Now()

	maxMinutes := uint(math.MaxInt64 / int64(time.Minute))
	if minutes > maxMinutes {
		return "", errs.ErrTokenValidityTooHigh
	}

	validity := time.Minute * time.Duration(minutes)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    baseUrl,
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(validity)),
			ID:        tokenID,
		},
		Role: role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	_, err := jwt.ParseWithClaims(
		tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}
			return []byte(secret), nil
		},
	)
	if err != nil {
		return claims, err
	}

	return claims, nil
}
