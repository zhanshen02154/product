package service

import (
	"context"
	"errors"
	"github.com/zhanshen02154/product/internal/application/dto"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/repository"
	"github.com/zhanshen02154/product/internal/domain/service"
	"github.com/zhanshen02154/product/internal/infrastructure/persistence/transaction"
	"github.com/zhanshen02154/product/pkg/swap"
	"github.com/zhanshen02154/product/proto/product"
)

type IProductApplicationService interface {
	AddProduct(ctx context.Context, productInfo *dto.ProductDto) (*dto.AddProductResponse, error)
	DeductInvetory(ctx context.Context, req *product.OrderDetailReq) error
}

type ProductApplicationService struct {
	productRepo          repository.IProductRepository
	productDomainService service.IProductDataService
	txManager            transaction.TransactionManager
}

func NewProductApplicationService(txManager transaction.TransactionManager, productRepo repository.IProductRepository) IProductApplicationService {
	return &ProductApplicationService{
		productDomainService: service.NewProductDataService(productRepo),
		productRepo:          productRepo,
		txManager:            txManager,
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

// DeductInvetory 扣减订单的库存
func (appService *ProductApplicationService) DeductInvetory(ctx context.Context, req *product.OrderDetailReq) error {
	return appService.txManager.ExecuteTransaction(ctx, func(txCtx context.Context) error {
		if len(req.Products) == 0 || len(req.Products) == 0 {
			return errors.New("没有数据")
		}
		return appService.productDomainService.DeductInvetory(ctx, req)
	})
}
