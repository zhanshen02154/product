package gorm

import (
	"context"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/repository"
	"gorm.io/gorm"
)

type InventoryStockChangeRecordRepositoryImpl struct {
	db *gorm.DB
}

// BatchCreate 批量创建库存变更记录
func (r *InventoryStockChangeRecordRepositoryImpl) BatchCreate(ctx context.Context, records []*model.InventoryStockChangeRecord) error {
	if len(records) == 0 {
		return nil
	}
	db := GetDBFromContext(ctx, r.db)
	return db.CreateInBatches(records, 100).Error
}

// NewInventoryStockChangeRecordRepository 创建库存变更记录仓储实例
func NewInventoryStockChangeRecordRepository(db *gorm.DB) repository.InventoryStockChangeRecordRepository {
	return &InventoryStockChangeRecordRepositoryImpl{db: db}
}
