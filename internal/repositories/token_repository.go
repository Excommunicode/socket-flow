package repositories

import (
	"context"
	stdErrors "errors"
	"fmt"
	"socket-flow/internal/jwt"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type TokenRepository interface {
	SaveJWToken(userID uuid.UUID, token string) error
	SaveJWTokens(userID uuid.UUID, accessToken, refreshToken string) error
	DeleteJWToken(userID uuid.UUID, token string) error
	DeleteJWTokens(userID uuid.UUID, accessToken, refreshToken string) error
	ValidateJWToken(userID uuid.UUID, token string) (bool, error)
}

type TokenRepositoryImpl struct {
	redisClient *redis.Client
}

func NewTokenRepository(client *redis.Client) *TokenRepositoryImpl {
	return &TokenRepositoryImpl{redisClient: client}
}

func (ts *TokenRepositoryImpl) SaveJWToken(userID uuid.UUID, token string) error {
	claims, err := jwt.ParseJWT(token)
	if err != nil {
		return errors.Wrap(err, "parse jwt")
	}
	duration := time.Until(claims.ExpiresAt.Time)
	if duration <= 0 {
		return stdErrors.New("token has already expired")
	}
	key := fmt.Sprintf("user:%d:jwt:%s", userID, claims.ID)
	return errors.Wrap(ts.redisClient.Set(context.Background(), key, token, duration).Err(), "save jwt to redis")
}

func (ts *TokenRepositoryImpl) SaveJWTokens(
	userID uuid.UUID, accessToken, refreshToken string,
) error {
	if err := ts.SaveJWToken(userID, accessToken); err != nil {
		return errors.Wrap(err, "save access token")
	}

	if err := ts.SaveJWToken(userID, refreshToken); err != nil {
		return errors.Wrap(err, "save refresh token")
	}

	return nil
}

func (ts *TokenRepositoryImpl) DeleteJWToken(userID uuid.UUID, token string) error {
	claims, err := jwt.ParseJWT(token)
	if err != nil {
		return errors.Wrap(err, "parse jwt")
	}

	key := fmt.Sprintf("user:%d:jwt:%s", userID, claims.ID)

	return errors.Wrap(ts.redisClient.Del(context.Background(), key).Err(), "delete jwt from redis")
}

func (ts *TokenRepositoryImpl) DeleteJWTokens(
	userID uuid.UUID, accessToken, refreshToken string,
) error {
	if err := ts.DeleteJWToken(userID, accessToken); err != nil {
		return errors.Wrap(err, "delete access token")
	}
	if err := ts.DeleteJWToken(userID, refreshToken); err != nil {
		return errors.Wrap(err, "delete refresh token")
	}
	return nil
}

func (ts *TokenRepositoryImpl) ValidateJWToken(userID uuid.UUID, token string) (bool, error) {
	claims, err := jwt.ParseJWT(token)
	if err != nil {
		return false, errors.Wrap(err, "parse jwt")
	}
	key := fmt.Sprintf("user:%d:jwt:%s", userID, claims.ID)
	storedToken, err := ts.redisClient.Get(context.Background(), key).Result()
	if err != nil {
		if stdErrors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, errors.Wrap(err, "get jwt from redis")
	}
	if storedToken == token {
		return true, nil
	}
	return false, nil
}
