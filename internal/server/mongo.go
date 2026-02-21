package server

import (
	"context"
	"fmt"
	"log/slog"

	"socket-flow/internal/config"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func InitMongoDB(ctx context.Context, cfg config.MongoConfig) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(cfg.URI)
	client, err := mongo.Connect(clientOptions)

	if err != nil {
		return nil, fmt.Errorf("cannot connect to mongo %+v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return client, fmt.Errorf("cannot ping mongo %+v", err)
	}

	slog.InfoContext(ctx, "mongodb connected", "uri", cfg.URI, "db", cfg.Database, "collection", cfg.Collection)
	return client, nil
}
