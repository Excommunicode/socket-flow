package server

import (
	"fmt"
	"socket-flow/internal/config"

	"github.com/jmoiron/sqlx"
)

func initDB(pgConfig config.PGConfig) (*sqlx.DB, error) {

	connect, err := sqlx.Open("pgx", pgConfig.DSN)
	if err != nil {
		return connect, fmt.Errorf("open db: %+v", err)
	}

	if err := connect.DB.Ping(); err != nil {
		return connect, fmt.Errorf("ping db: %v", err)
	}

	return connect, nil
}
