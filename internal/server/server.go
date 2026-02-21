package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"socket-flow/internal/config"
	"socket-flow/pkg/postgres"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewServer(ctx context.Context) (*http.Server, error) {

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	var (
		db          *sqlx.DB
		mongoClient *mongo.Client
		redisClient *redis.Client
	)

	defer func() {
		if err != nil {
			if closeErr := closeResources(ctx, db, mongoClient, redisClient); closeErr != nil {
				slog.Error("failed to cleanup resources during server init failure", "err", closeErr)
			}
		}
	}()

	db, err = initDB(cfg.Postgres)
	if err != nil {
		return nil, err
	}

	mongoClient, err = InitMongoDB(ctx, cfg.Mongo)
	if err != nil {
		return nil, err
	}

	redisClient, err = InitRedis(cfg.Redis)
	if err != nil {
		return nil, err
	}

	upgrader := InitWebSocket(cfg.WebSocket)

	pgClient := postgres.NewClient(db)
	repositories := InitRepositories(*pgClient, mongoClient, redisClient, cfg.Mongo)
	services := InitServices(pgClient, repositories)
	handler := InitHandler(services, upgrader)
	routers := InitRouters(handler)

	if err := runPgMigrations(cfg.Postgres); err != nil {
		return nil, fmt.Errorf("pg migrations failed: %+v", err)
	}
	if err := runMongoMigration(cfg.Mongo); err != nil {
		return nil, fmt.Errorf("mongo migrations failed: %+v", err)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           routers,
		ReadHeaderTimeout: 5 * time.Second,
	}

	srv.RegisterOnShutdown(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := closeResources(ctx, db, mongoClient, redisClient); err != nil {
			slog.Error("error shutting down resources", "err", err)
		}
	})

	return srv, nil
}

func closeResources(ctx context.Context, db *sqlx.DB, mongoClient *mongo.Client, redisClient *redis.Client) error {
	var errs []error

	if db != nil {
		if err := db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close db: %w", err))
		}
	}

	if mongoClient != nil {
		if err := mongoClient.Disconnect(ctx); err != nil {
			errs = append(errs, fmt.Errorf("disconnect mongo: %w", err))
		}
	}

	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close redis: %w", err))
		}
	}
	return errors.Join(errs...)
}
