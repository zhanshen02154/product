package repository

import (
	"context"
	"github.com/zhanshen02154/product/internal/domain/model"
)

type IProductRepository interface {
	FindProductByID(ctx context.Context, id int64) (*model.Product, error)
	CreateProduct(ctx context.Context, productInfo *model.Product) (int64, error)
}
