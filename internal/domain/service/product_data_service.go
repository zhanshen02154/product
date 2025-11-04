package service

import (
	"context"
	"errors"
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
	if len(req.OrderDetail) == 0 {
		return errors.New("没有库存内容")
	}
	var productIds []int64
	productSizeIds := make([]int64, len(req.OrderDetail))
	for _, item := range req.OrderDetail {
		productIds = append(productIds, item.ProductId)
		productSizeIds = append(productSizeIds, item.ProductSizeId)
	}
	productSizeList, err := u.productRepository.FindProductSizeListByIds(ctx, productSizeIds)
	if err != nil {
		return err
	}
	if len(productSizeList) == 0 {
		return errors.New("找不到产品规格数据")
	}
	productList, err := u.productRepository.FindProductListByIds(ctx, productIds)
	if err != nil {
		return err
	}
	if len(productList) == 0 {
		return errors.New("找不到产品数据")
	}

	return nil
}
