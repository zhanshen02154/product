package transaction

import "context"

type TransactionManager interface {
	Execute(ctx context.Context, fn func(txCtx context.Context) error) error
	ExecuteWithBarrier(ctx context.Context, fn func(txCtx context.Context) error) error
}
