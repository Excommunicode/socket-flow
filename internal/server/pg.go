package server

import (
	"context"
	"socket-flow/internal/config"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func initDB(ctx context.Context, pgConfig config.PGConfig) (*sqlx.DB, error) {
	connect, err := sqlx.Open("pgx", pgConfig.DSN())
	if err != nil {
		return connect, errors.Wrap(err, "open db")
	}

	err = connect.PingContext(ctx)

	if err != nil {
		return connect, errors.Wrap(err, "ping db")
	}

	return connect, nil
}
