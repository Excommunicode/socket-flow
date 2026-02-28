package postgres

import "context"

type Transactor interface {
	WithinROTransaction(context.Context, func(ctx context.Context) error) error
	WithinNewROTransaction(context.Context, func(ctx context.Context) error) error
	WithinRWTransaction(context.Context, func(ctx context.Context) error) error
	WithinNewRWTransaction(context.Context, func(ctx context.Context) error) error
}
