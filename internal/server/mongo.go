package server

import (
	"context"
	"fmt"
	"log/slog"

	"socket-flow/internal/config"
	"socket-flow/internal/errors"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func initMongoDB(ctx context.Context, cfg config.MongoConfig) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(cfg.URI)
	client, err := mongo.Connect(clientOptions)

	if err != nil {
		return nil, fmt.Errorf("%w: %w", errors.ErrMongoPingConnection, err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return client, fmt.Errorf("%w: %w", errors.ErrMongoPingConnection, err)
	}

	slog.InfoContext(ctx, "mongodb connected", "uri", cfg.URI, "db", cfg.Database, "collection", cfg.Collection)

	return client, nil
}
