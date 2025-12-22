package gorm

import (
	"context"
	"database/sql"
	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/zhanshen02154/product/internal/infrastructure/persistence/transaction"
	"gorm.io/gorm"
)

type GormTransactionManager struct {
	db *gorm.DB
}

type txKey struct{}

func (gtm *GormTransactionManager) Execute(ctx context.Context, fn func(txCtx context.Context) error) error {
	// 使用 NewDB session 为每次事务创建独立会话，降低 statement 缓存/复用时的互斥争用
	session := gtm.db.Session(&gorm.Session{NewDB: true})
	return session.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}

func NewGormTransactionManager(db *gorm.DB) transaction.TransactionManager {
	return &GormTransactionManager{db: db}
}

// GetDBFromContext 从 context 中提取 DB 实例（事务或非事务）
func GetDBFromContext(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx // 返回事务实例
	}
	return defaultDB.WithContext(ctx) // 返回默认的非事务实例
}

// ExecuteWithBarrier 开启子事务屏障的事务
func (gtm *GormTransactionManager) ExecuteWithBarrier(ctx context.Context, fn func(txCtx context.Context) error) error {
	barrier, err := dtmgrpc.BarrierFromGrpc(ctx)
	if err != nil {
		return err
	}
	sqlDb, err := gtm.db.DB()
	if err != nil {
		return err
	}
	session := gtm.db.Session(&gorm.Session{
		SkipHooks:                false,
		SkipDefaultTransaction:   true,
		DisableNestedTransaction: true,
		Context:                  ctx,
		CreateBatchSize:          2000,
	})
	return barrier.CallWithDB(sqlDb, func(tx1 *sql.Tx) error {
		session.Statement.ConnPool = tx1
		txCtx := context.WithValue(ctx, txKey{}, session)
		return fn(txCtx)
	})
}
