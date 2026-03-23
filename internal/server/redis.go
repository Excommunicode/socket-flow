package server

import (
	"context"
	"socket-flow/internal/config"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

func initRedis(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return client, errors.Wrap(err, "failed to connect to redis")
	}

	return client, nil
}
