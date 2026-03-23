package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type PgClient struct {
	db *sqlx.DB
}

func NewClient(db *sqlx.DB) *PgClient {
	return &PgClient{
		db: db,
	}
}

func (c *PgClient) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if tx, ok := TxFromContext(ctx); ok && tx != nil {
		res, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, errors.Wrap(err, "exec query in transaction")
		}

		return res, nil
	}

	res, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "exec query")
	}

	return res, nil
}

func (c *PgClient) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if tx, ok := TxFromContext(ctx); ok && tx != nil {
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, errors.Wrap(err, "query rows in transaction")
		}

		return rows, nil
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "query rows")
	}

	return rows, nil
}

func (c *PgClient) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	if tx, ok := TxFromContext(ctx); ok && tx != nil {
		return tx.QueryRowContext(ctx, query, args...)
	}

	return c.db.QueryRowContext(ctx, query, args...)
}

func (c *PgClient) Select(ctx context.Context, dest any, query string, args ...any) error {
	if tx, ok := TxFromContext(ctx); ok && tx != nil {
		err := tx.SelectContext(ctx, dest, query, args...)
		if err != nil {
			return errors.Wrap(err, "select in transaction")
		}

		return nil
	}

	err := c.db.SelectContext(ctx, dest, query, args...)
	if err != nil {
		return errors.Wrap(err, "select")
	}

	return nil
}
