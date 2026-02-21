package server

import (
	"context"
	"fmt"
	"socket-flow/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitRedis(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return client, fmt.Errorf("failed to connect to redis: %+v", err)
	}
	return client, nil
}
