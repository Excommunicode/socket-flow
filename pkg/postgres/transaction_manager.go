package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

var (
	txRO = &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}
	txRW = &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	}
)

type txKey struct{}

type TransactionManager struct {
	db *sqlx.DB
}

func NewTransactionManager(db *sqlx.DB) Transactor {
	return &TransactionManager{db: db}
}

func (t *TransactionManager) WithinROTransaction(ctx context.Context, f func(ctx context.Context) error) error {
	if _, ok := TxFromContext(ctx); ok {
		return f(ctx)
	}
	return t.withTx(ctx, txRO, f)
}

func (t *TransactionManager) WithinNewROTransaction(ctx context.Context, f func(ctx context.Context) error) error {
	return t.withTx(ctx, txRO, f)
}

func (t *TransactionManager) WithinRWTransaction(ctx context.Context, f func(ctx context.Context) error) error {
	if _, ok := TxFromContext(ctx); ok {
		return f(ctx)
	}
	return t.withTx(ctx, txRW, f)
}

func (t *TransactionManager) WithinNewRWTransaction(ctx context.Context, f func(ctx context.Context) error) error {
	return t.withTx(ctx, txRW, f)
}

func (t *TransactionManager) withTx(ctx context.Context, opts *sql.TxOptions, f func(ctx context.Context) error) error {
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

	ctxWithTx := contextWithTx(ctx, tx)
	err = f(ctxWithTx)
	return err
}

func TxFromContext(ctx context.Context) (*sqlx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sqlx.Tx)
	return tx, ok
}

func contextWithTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}
