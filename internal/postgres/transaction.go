package postgres

import "context"

type Transactor interface {
	WithinROTransaction(ctx context.Context, f func(ctx context.Context) error) error
	WithinRWTransaction(ctx context.Context, f func(ctx context.Context) error) error
}
