package gorm

import (
	"context"
	"github.com/zhanshen02154/product/internal/infrastructure/persistence/transaction"
	"gorm.io/gorm"
)

type GormTransactionManager struct {
	db *gorm.DB
}

type txKey struct{}

func (gtm *GormTransactionManager) ExecuteTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	tx := gtm.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)
	err := fn(txCtx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func NewGormTransactionManager(db *gorm.DB) transaction.TransactionManager {
	return &GormTransactionManager{db: db}
}

// GetDBFromContext 从 context 中提取 DB 实例（事务或非事务）
func GetDBFromContext(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx // 返回事务实例
	}
	return defaultDB // 返回默认的非事务实例
}
