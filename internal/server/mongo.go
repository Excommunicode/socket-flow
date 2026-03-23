package server

import (
	"context"
	"log/slog"

	"socket-flow/internal/config"
	appErrors "socket-flow/internal/errors"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func initMongoDB(ctx context.Context, cfg config.MongoConfig) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(cfg.URI)
	client, err := mongo.Connect(clientOptions)

	if err != nil {
		return nil, errors.Wrap(err, appErrors.ErrMongoPingConnection.Error())
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return client, errors.Wrap(err, appErrors.ErrMongoPingConnection.Error())
	}

	slog.InfoContext(ctx, "mongodb connected", "uri", cfg.URI, "db", cfg.Database, "collection", cfg.Collection)

	return client, nil
}
