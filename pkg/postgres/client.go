package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/jmoiron/sqlx"
)

type Client struct {
	Db *sqlx.DB
}

func NewClient(db *sqlx.DB) *Client {
	return &Client{
		Db: db,
	}
}

var savepointCounter atomic.Uint64

// ErrNoTransactionInContext возвращается, если в контексте нет текущей транзакции.
var ErrNoTransactionInContext = errors.New("no transaction in context")

type txContextKey struct{}

var txKey = &txContextKey{}

// WithTx сохраняет транзакцию в контексте.
func WithTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

// TxFromContext возвращает текущую транзакцию из контекста или nil.
func TxFromContext(ctx context.Context) *sqlx.Tx {
	v := ctx.Value(txKey)
	if v == nil {
		return nil
	}
	tx, _ := v.(*sqlx.Tx)
	return tx
}

func (m *Client) WithReadOnlyTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return WithReadOnlyTransaction(ctx, m.Db, fn)
}

func (m *Client) WithReadWriteTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return WithReadWriteTransaction(ctx, m.Db, fn)
}

func (m *Client) WithNestedReadOnly(ctx context.Context, fn func(ctx context.Context) error) error {
	return WithNestedReadOnly(ctx, fn)
}

func (m *Client) WithNestedReadWrite(ctx context.Context, fn func(ctx context.Context) error) error {
	return WithNestedReadWrite(ctx, fn)
}

func WithReadOnlyTransaction(ctx context.Context, db *sqlx.DB, fn func(ctx context.Context) error) error {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
		}
	}()
	txCtx := WithTx(ctx, tx)
	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Join(err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

func WithReadWriteTransaction(ctx context.Context, db *sqlx.DB, fn func(ctx context.Context) error) error {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{ReadOnly: false, Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
		}
	}()
	txCtx := WithTx(ctx, tx)
	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Join(err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

func WithNestedReadOnly(ctx context.Context, fn func(ctx context.Context) error) error {
	parent := TxFromContext(ctx)
	if parent == nil {
		return ErrNoTransactionInContext
	}
	name := fmt.Sprintf("sp_%d", savepointCounter.Add(1))
	if _, err := parent.ExecContext(ctx, "SET LOCAL transaction_read_only = on"); err != nil {
		return err
	}
	if _, err := parent.ExecContext(ctx, "SAVEPOINT "+name); err != nil {
		_, _ = parent.ExecContext(ctx, "SET LOCAL transaction_read_only = off")
		return err
	}
	var fnErr error
	defer func() {
		if p := recover(); p != nil {
			_, _ = parent.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+name)
			_, _ = parent.ExecContext(ctx, "SET LOCAL transaction_read_only = off")
			panic(p)
		}
		if fnErr != nil {
			_, _ = parent.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+name)
		} else {
			_, _ = parent.ExecContext(ctx, "RELEASE SAVEPOINT "+name)
		}
		_, _ = parent.ExecContext(ctx, "SET LOCAL transaction_read_only = off")
	}()
	fnErr = fn(ctx)
	return fnErr
}

func WithNestedReadWrite(ctx context.Context, fn func(ctx context.Context) error) error {
	parent := TxFromContext(ctx)
	if parent == nil {
		return ErrNoTransactionInContext
	}
	name := fmt.Sprintf("sp_%d", savepointCounter.Add(1))
	if _, err := parent.ExecContext(ctx, "SAVEPOINT "+name); err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_, _ = parent.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+name)
			panic(p)
		}
	}()
	if err := fn(ctx); err != nil {
		if _, rbErr := parent.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+name); rbErr != nil {
			return errors.Join(err, rbErr)
		}
		return err
	}
	if _, err := parent.ExecContext(ctx, "RELEASE SAVEPOINT "+name); err != nil {
		return err
	}
	return nil
}
