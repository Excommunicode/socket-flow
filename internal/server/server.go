package server

import (
	"context"
	stdErrors "errors"
	"log/slog"
	"net/http"
	"socket-flow/internal/postgres"
	"time"

	"socket-flow/internal/config"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewServer(ctx context.Context) (*http.Server, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, errors.Wrap(err, "load config")
	}

	var (
		db          *sqlx.DB
		mongoClient *mongo.Client
		redisClient *redis.Client
	)

	defer func() {
		if err != nil {
			err := closeResources(ctx, db, mongoClient, redisClient)
			if err != nil {
				slog.Error("failed to cleanup resources during server init failure", "err", err)
			}
		}
	}()

	db, err = initDB(ctx, cfg.Postgres)
	if err != nil {
		return nil, errors.Wrap(err, "init postgres")
	}

	mongoClient, err = initMongoDB(ctx, cfg.Mongo)
	if err != nil {
		return nil, errors.Wrap(err, "init mongo")
	}

	redisClient, err = initRedis(cfg.Redis)
	if err != nil {
		return nil, errors.Wrap(err, "init redis")
	}

	err = runPgMigrations(cfg.Postgres)

	if err != nil {
		return nil, errors.Wrap(err, "pg migrations failed")
	}

	err = runMongoMigration(cfg.Mongo)

	if err != nil {
		return nil, errors.Wrap(err, "mongo migrations failed")
	}

	transactor := postgres.NewTransactionManager(db)
	upgrader := InitWebSocket(cfg.WebSocket)
	pgClient := postgres.NewClient(db)
	repositories := InitRepositories(pgClient, mongoClient, redisClient, cfg.Mongo)
	services, err := initServices(transactor, repositories, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "init services")
	}

	schedulerCtx, schedulerCancel := context.WithCancel(ctx)
	go func() {
		schedulerErr := services.MessageScheduler.StartCleanupScheduler(schedulerCtx)
		if schedulerErr != nil && !stdErrors.Is(schedulerErr, context.Canceled) {
			slog.Error("message cleanup scheduler stopped with error", "err", schedulerErr)
		}
	}()

	handler := initHandler(services, upgrader)
	routers := initRouters(handler)

	srv := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           routers,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	srv.RegisterOnShutdown(func() {
		schedulerCancel()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if closeErr := closeResources(shutdownCtx, db, mongoClient, redisClient); closeErr != nil {
			slog.Error("error shutting down resources", "err", closeErr)
		}
	})

	return srv, nil
}

func closeResources(ctx context.Context, db *sqlx.DB, mongoClient *mongo.Client, redisClient *redis.Client) error {
	var errs []error

	if db != nil {
		if err := db.Close(); err != nil {
			errs = append(errs, errors.Wrap(err, "close db"))
		}
	}

	if mongoClient != nil {
		if err := mongoClient.Disconnect(ctx); err != nil {
			errs = append(errs, errors.Wrap(err, "disconnect mongo"))
		}
	}

	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			errs = append(errs, errors.Wrap(err, "close redis"))
		}
	}
	return stdErrors.Join(errs...)
}
