package repository

import (
	"context"
	"github.com/zhanshen02154/product/internal/domain/model"
)

// OrderInventoryEventRepository 订单库存仓储服务
type OrderInventoryEventRepository interface {
	// FindEventExistsByOrderId 检查订单ID是否已经被处理过
	FindEventExistsByOrderId(ctx context.Context, orderId int64) (bool, error)
	Create(ctx context.Context, eventInfo *model.OrderInventoryEvent) (*model.OrderInventoryEvent, error)
}
