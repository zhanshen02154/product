package service

import (
	"context"
	"github.com/zhanshen02154/product/internal/application/dto"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/repository"
	"github.com/zhanshen02154/product/internal/domain/service"
	"github.com/zhanshen02154/product/pkg/swap"
)

type IProductApplicationService interface {
	AddProduct(ctx context.Context, productInfo *dto.ProductDto) (*dto.AddProductResponse, error)
}

type ProductApplicationService struct {
	productRepo          repository.IProductRepository
	productDomainService service.IProductDataService
}

func NewProductApplicationService(productRepo repository.IProductRepository) IProductApplicationService {
	return &ProductApplicationService{
		productDomainService: service.NewProductDataService(productRepo),
		productRepo:          productRepo,
	}
}

// AddProduct 添加产品
func (appService *ProductApplicationService) AddProduct(ctx context.Context, productInfo *dto.ProductDto) (*dto.AddProductResponse, error) {
	productModel := &model.Product{}
	if err := swap.SwapTo(productInfo, productModel); err != nil {
		return nil, err
	}
	productId, err := appService.productDomainService.AddProduct(productModel)
	if err != nil {
		return nil, err
	}
	return &dto.AddProductResponse{Id: productId}, nil
}
