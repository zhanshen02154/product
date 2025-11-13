package transaction

import "context"

type TransactionManager interface {
	ExecuteTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}
