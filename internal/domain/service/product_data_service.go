package service

import (
	"context"
	"errors"
	"github.com/zhanshen02154/product/internal/application/dto"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/repository"
)

type IProductDataService interface {
	AddProduct(ctx context.Context, productInfo *model.Product) (int64, error)
	DeductOrderInvetory(ctx context.Context, req *dto.OrderProductInvetoryDto) error
	DeductOrderInvetoryRevert(ctx context.Context, req *dto.OrderProductInvetoryDto) error
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

// DeductOrderInvetory 扣减库存
func (u *ProductDataService) DeductOrderInvetory(ctx context.Context, req *dto.OrderProductInvetoryDto) error {
	var err error
	for _, item := range req.ProductInvetory {
		err = u.productRepository.DeductProductInvetory(ctx, item.Id, item.Count)
		if err != nil {
			break
		}
	}
	if err != nil {
		return err
	}
	for _, item := range req.ProductSizeInvetory {
		err = u.productRepository.DeductProductSizeInvetory(ctx, item.Id, item.Count)
		if err != nil {
			break
		}
	}
	return err
}

// DeductOrderInvetoryRevert 扣减订单库存补偿操作
func (u *ProductDataService) DeductOrderInvetoryRevert(ctx context.Context, req *dto.OrderProductInvetoryDto) error {
	var err error
	if len(req.ProductSizeInvetory) == 0 || len(req.ProductSizeInvetory) == 0 {
		err = errors.New("product data cannot be empty")
	}
	for _, item := range req.ProductInvetory {
		err = u.productRepository.DeductProductInvetoryRevert(ctx, item.Id, item.Count)
		if err != nil {
			break
		}
	}
	if err != nil {
		return err
	}
	for _, item := range req.ProductSizeInvetory {
		err = u.productRepository.DeductProductSizeInvetoryRevert(ctx, item.Id, item.Count)
		if err != nil {
			break
		}
	}
	return err
}
