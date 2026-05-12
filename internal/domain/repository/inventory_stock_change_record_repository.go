package repository

import (
	"context"
	"github.com/zhanshen02154/product/internal/domain/model"
)

// InventoryStockChangeRecordRepository 库存变更记录仓储接口
type InventoryStockChangeRecordRepository interface {
	// BatchCreate 批量创建库存变更记录
	BatchCreate(ctx context.Context, records []*model.InventoryStockChangeRecord) error
	// GetSalesVolume 获取SKU在指定时间范围内的销量（订单支付扣减库存的数量总和）和平均销量
	GetSalesVolume(ctx context.Context, skuID int64, startTime, endTime string) (int64, float64, error)
}
