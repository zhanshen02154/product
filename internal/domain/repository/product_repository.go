package repository

import (
	"context"
	"github.com/zhanshen02154/product/internal/domain/model"
)

type IProductRepository interface {
	FindProductByID(ctx context.Context, id int64) (*model.Product, error)
	CreateProduct(ctx context.Context, productInfo *model.Product) (int64, error)
	FindProductSizeListByIds(ctx context.Context, ids []int64) ([]model.ProductSize, error)
	FindProductListByIds(ctx context.Context, productIds []int64) ([]model.Product, error)
	DeductProductInvetory(ctx context.Context, id int64, num int64) error
	DeductProductSizeInvetory(ctx context.Context, id int64, num int64) error
}
