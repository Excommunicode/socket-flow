package postgres

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

var queryBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

type Client struct {
	pg           *sqlx.DB
	QueryBuilder squirrel.StatementBuilderType
}

func NewClient(db *sqlx.DB) *Client {
	return &Client{
		pg:           db,
		QueryBuilder: queryBuilder,
	}
}

func (c *Client) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if tx, ok := TxFromContext(ctx); ok && tx != nil {
		return tx.ExecContext(ctx, query, args...)
	}
	return c.pg.ExecContext(ctx, query, args...)
}

func (c *Client) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if tx, ok := TxFromContext(ctx); ok {
		return tx.QueryContext(ctx, query, args...)
	}
	return c.pg.QueryContext(ctx, query, args...)
}

func (c *Client) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if tx, ok := TxFromContext(ctx); ok && tx != nil {
		return tx.QueryRowContext(ctx, query, args...)
	}

	return c.pg.QueryRowContext(ctx, query, args...)
}

func (c *Client) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	if tx, ok := TxFromContext(ctx); ok && tx != nil {
		return tx.SelectContext(ctx, dest, query, args...)
	}
	return c.pg.SelectContext(ctx, dest, query, args...)
}
