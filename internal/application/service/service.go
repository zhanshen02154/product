package service

import (
	"context"
	"github.com/zhanshen02154/product/internal/application/dto"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/repository"
	"github.com/zhanshen02154/product/internal/domain/service"
	gorm2 "github.com/zhanshen02154/product/internal/infrastructure/persistence/gorm"
	"github.com/zhanshen02154/product/internal/infrastructure/persistence/transaction"
	"github.com/zhanshen02154/product/pkg/swap"
	"gorm.io/gorm"
)

type IProductApplicationService interface {
	AddProduct(ctx context.Context, productInfo *dto.ProductDto) (*dto.AddProductResponse, error)
}

type ProductApplicationService struct {
	productRepo          repository.IProductRepository
	productDomainService service.IProductDataService
	txManager            transaction.TransactionManager
}

func NewProductApplicationService(db *gorm.DB) IProductApplicationService {
	productRepo := gorm2.NewProductRepository(db)
	return &ProductApplicationService{
		productDomainService: service.NewProductDataService(productRepo),
		productRepo:          productRepo,
		txManager:            gorm2.NewGormTransactionManager(db),
	}
}

// AddProduct 添加产品
func (appService *ProductApplicationService) AddProduct(ctx context.Context, productInfo *dto.ProductDto) (*dto.AddProductResponse, error) {
	productModel := &model.Product{}
	if err := swap.SwapTo(productInfo, productModel); err != nil {
		return nil, err
	}
	productId, err := appService.productDomainService.AddProduct(ctx, productModel)
	if err != nil {
		return nil, err
	}
	return &dto.AddProductResponse{Id: productId}, nil
}
