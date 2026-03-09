package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type txKey struct{}

type TransactionManager struct {
	db *sqlx.DB
}

func NewTransactionManager(db *sqlx.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

func (t *TransactionManager) WithinROTransaction(ctx context.Context, f func(ctx context.Context) error) error {
	return t.withTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true}, f)
}

func (t *TransactionManager) WithinRWTransaction(ctx context.Context, f func(ctx context.Context) error) error {
	return t.withTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false}, f)
}

func (t *TransactionManager) withTx(ctx context.Context,
	opts *sql.TxOptions, fn func(ctx context.Context) error) (err error) {

	if _, ok := TxFromContext(ctx); ok {
		return fn(ctx)
	}

	tx, err := t.db.BeginTxx(ctx, opts)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()

			panic(p)
		}

		if err != nil {
			_ = tx.Rollback()

			return
		}

		if cErr := tx.Commit(); cErr != nil {
			err = fmt.Errorf("commit tx: %w", cErr)
		}
	}()

	err = fn(contextWithTx(ctx, tx))

	return err
}

func TxFromContext(ctx context.Context) (*sqlx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sqlx.Tx)

	return tx, ok
}

func contextWithTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}
