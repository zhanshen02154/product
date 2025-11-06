package service

import (
	"context"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/repository"
	"github.com/zhanshen02154/product/proto/product"
)

type IProductDataService interface {
	AddProduct(ctx context.Context, productInfo *model.Product) (int64, error)
	DeductInvetory(ctx context.Context, req *product.OrderDetailReq) error
}

// NewProductDataService 创建
func NewProductDataService(productRepository repository.IProductRepository) IProductDataService {
	return &ProductDataService{productRepository}
}

type ProductDataService struct {
	productRepository repository.IProductRepository
}

// AddProduct 插入
func (u *ProductDataService) AddProduct(ctx context.Context, product *model.Product) (int64, error) {
	return u.productRepository.CreateProduct(ctx, product)
}

// DeductInvetory 扣减库存
func (u *ProductDataService) DeductInvetory(ctx context.Context, req *product.OrderDetailReq) error {
	var err error
	for _, item := range req.Products {
		err = u.productRepository.DeductProductInvetory(ctx, item.ProductId, item.ProductNum)
		if err != nil {
			break
		}
		err = u.productRepository.DeductProductSizeInvetory(ctx, item.ProductSizeId, item.ProductNum)
		if err != nil {
			break
		}
	}
	return err
}
