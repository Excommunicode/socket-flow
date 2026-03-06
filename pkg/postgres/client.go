package postgres

import (
	"context"
	"database/sql"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type Client struct {
	db           *sqlx.DB
	QueryBuilder sq.StatementBuilderType
}

func NewClient(db *sqlx.DB) *Client {
	return &Client{
		db:           db,
		QueryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (c *Client) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if tx, ok := TxFromContext(ctx); ok && tx != nil {
		return tx.ExecContext(ctx, query, args...)
	}

	return c.db.ExecContext(ctx, query, args...)
}

func (c *Client) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if tx, ok := TxFromContext(ctx); ok && tx != nil {
		return tx.QueryContext(ctx, query, args...)
	}

	return c.db.QueryContext(ctx, query, args...)
}

func (c *Client) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	if tx, ok := TxFromContext(ctx); ok && tx != nil {
		return tx.QueryRowContext(ctx, query, args...)
	}

	return c.db.QueryRowContext(ctx, query, args...)
}

func (c *Client) Select(ctx context.Context, dest any, query string, args ...any) error {
	if tx, ok := TxFromContext(ctx); ok && tx != nil {
		return tx.SelectContext(ctx, dest, query, args...)
	}

	return c.db.SelectContext(ctx, dest, query, args...)
}

func (c *Client) ToSQL(q sq.Sqlizer) (string, []any, error) {
	query, args, err := q.ToSql()
	slog.Debug("sql", "query", query, "args", args, "err", err)

	return query, args, err
}
