package postgres

import "context"

type TransactionManager interface {
	WithReadOnlyTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	WithReadWriteTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	WithNestedReadOnly(ctx context.Context, fn func(ctx context.Context) error) error
	WithNestedReadWrite(ctx context.Context, fn func(ctx context.Context) error) error
}
