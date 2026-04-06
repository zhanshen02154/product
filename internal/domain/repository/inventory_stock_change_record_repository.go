package repository

import (
	"context"
	"github.com/zhanshen02154/product/internal/domain/model"
)

// InventoryStockChangeRecordRepository 库存变更记录仓储接口
type InventoryStockChangeRecordRepository interface {
	// BatchCreate 批量创建库存变更记录
	BatchCreate(ctx context.Context, records []*model.InventoryStockChangeRecord) error
}
