package repositories

import (
	"context"
	"errors"
	"fmt"
	"socket-flow/pkg/jwt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type TokenRepository interface {
	SaveJWToken(userID uuid.UUID, token string) error
	SaveJWTokens(userID uuid.UUID, accessToken, refreshToken string) error
	DeleteJWToken(userID uuid.UUID, token string) error
	DeleteJWTokens(userID uuid.UUID, accessToken, refreshToken string) error
	ValidateJWToken(userID uuid.UUID, token string) (bool, error)
}

type tokenRepository struct {
	redisClient *redis.Client
}

func NewTokenRepository(client *redis.Client) TokenRepository {
	return &tokenRepository{redisClient: client}
}

func (ts *tokenRepository) SaveJWToken(userID uuid.UUID, token string) error {
	claims, err := jwt.ParseJWT(token)
	if err != nil {
		return err
	}
	duration := time.Until(claims.ExpiresAt.Time)
	if duration <= 0 {
		return errors.New("token has already expired")
	}
	key := fmt.Sprintf("user:%d:jwt:%s", userID, claims.ID)
	return ts.redisClient.Set(context.Background(), key, token, duration).Err()
}

func (ts *tokenRepository) SaveJWTokens(
	userID uuid.UUID, accessToken, refreshToken string,
) error {
	if err := ts.SaveJWToken(userID, accessToken); err != nil {
		return err
	}
	if err := ts.SaveJWToken(userID, refreshToken); err != nil {
		return err
	}
	return nil
}

func (ts *tokenRepository) DeleteJWToken(userID uuid.UUID, token string) error {
	claims, err := jwt.ParseJWT(token)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("user:%d:jwt:%s", userID, claims.ID)
	return ts.redisClient.Del(context.Background(), key).Err()
}

func (ts *tokenRepository) DeleteJWTokens(
	userID uuid.UUID, accessToken, refreshToken string,
) error {
	if err := ts.DeleteJWToken(userID, accessToken); err != nil {
		return err
	}
	if err := ts.DeleteJWToken(userID, refreshToken); err != nil {
		return err
	}
	return nil
}

func (ts *tokenRepository) ValidateJWToken(userID uuid.UUID, token string) (bool, error) {
	claims, err := jwt.ParseJWT(token)
	if err != nil {
		return false, err
	}
	key := fmt.Sprintf("user:%d:jwt:%s", userID, claims.ID)
	storedToken, err := ts.redisClient.Get(context.Background(), key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	if storedToken == token {
		return true, nil
	}
	return false, nil
}
