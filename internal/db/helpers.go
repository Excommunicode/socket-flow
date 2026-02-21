package db

import (
	"context"

	"socket-flow/pkg/postgres"

	"github.com/jmoiron/sqlx"
)

func NewQueriesFromDB(db *sqlx.DB) *Queries {
	return New(db)
}

func NewQueriesFromTx(tx *sqlx.Tx) *Queries {
	return New(tx)
}

func QueriesFromContext(ctx context.Context, db *sqlx.DB) *Queries {
	tx := postgres.TxFromContext(ctx)
	if tx != nil {
		return NewQueriesFromTx(tx)
	}
	return NewQueriesFromDB(db)
}
