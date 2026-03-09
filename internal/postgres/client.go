package postgres

import (
	"context"
	"database/sql"
)

type Client interface {
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
	Select(ctx context.Context, dest any, query string, args ...any) error
}
