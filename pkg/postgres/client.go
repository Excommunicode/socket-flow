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

func (c *Client) Exec(query string, args ...interface{}) (sql.Result, error) {
	res, err := c.pg.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := c.pg.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (c *Client) QueryRow(query string, args ...interface{}) *sql.Row {
	return c.pg.QueryRow(query, args...)
}

func (c *Client) Select(dest interface{}, query string, args ...interface{}) error {
	return c.pg.Select(dest, query, args...)
}

func (c *Client) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return c.pg.SelectContext(ctx, dest, query, args...)
}
