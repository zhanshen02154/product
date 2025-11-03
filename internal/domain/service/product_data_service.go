package service

import (
	"context"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/repository"
)

type IProductDataService interface {
	AddProduct(ctx context.Context, productInfo *model.Product) (int64, error)
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
