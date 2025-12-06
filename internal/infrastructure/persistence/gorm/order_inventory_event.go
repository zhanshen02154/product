package gorm

import (
	"context"
	"errors"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/repository"
	"gorm.io/gorm"
)

type OrderInventoryEventRepositoryImpl struct {
	db *gorm.DB
}

// FindEventExistsByOrderId 检查订单ID是否已经被处理过
func (r *OrderInventoryEventRepositoryImpl) FindEventExistsByOrderId(ctx context.Context, orderId int64) (bool, error) {
	eventExists := &model.OrderInventoryEvent{}
	err := r.db.WithContext(ctx).Debug().Model(eventExists).Where("order_id = ?", orderId).Select("order_id").First(eventExists).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

// Create 创建事件记录
func (r *OrderInventoryEventRepositoryImpl) Create(ctx context.Context, eventInfo *model.OrderInventoryEvent) (*model.OrderInventoryEvent, error) {
	db := GetDBFromContext(ctx, r.db)
	err := db.Model(eventInfo).Create(eventInfo).Error
	if err != nil {
		return nil, err
	}
	return eventInfo, nil
}

// NewOrderInventoryEventRepositoryImpl 初始化
func NewOrderInventoryEventRepositoryImpl(db *gorm.DB) repository.OrderInventoryEventRepository {
	return &OrderInventoryEventRepositoryImpl{db: db}
}
