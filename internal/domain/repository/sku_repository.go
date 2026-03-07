package repository

import (
	"context"
	"github.com/zhanshen02154/product/internal/domain/model"
)

type ProductSkuRepository interface {
	BatchGetSkuByIDsWithFields(ctx context.Context, skuIDs []int64) ([]model.ProductSku, error)
	DeductInventoryById(ctx context.Context, id int64, count uint32) error
}
