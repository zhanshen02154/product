package service

import (
	"context"
	"errors"
	"github.com/zhanshen02154/product/internal/application/dto"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/service"
	"github.com/zhanshen02154/product/internal/infrastructure"
	"github.com/zhanshen02154/product/pkg/swap"
	"github.com/zhanshen02154/product/proto/product"
)

type IProductApplicationService interface {
	AddProduct(ctx context.Context, productInfo *dto.ProductDto) (*dto.AddProductResponse, error)
	DeductInvetory(ctx context.Context, req *product.OrderDetailReq) error
}

type ProductApplicationService struct {
	productDomainService service.IProductDataService
	serviceContext       *infrastructure.ServiceContext
}

func NewProductApplicationService(serviceContext *infrastructure.ServiceContext) IProductApplicationService {
	return &ProductApplicationService{
		productDomainService: service.NewProductDataService(serviceContext.OrderRepository),
		serviceContext:       serviceContext,
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
	return appService.serviceContext.TxManager.ExecuteTransaction(ctx, func(txCtx context.Context) error {
		if len(req.Products) == 0 || len(req.Products) == 0 {
			return errors.New("没有数据")
		}
		return appService.productDomainService.DeductInvetory(ctx, req)
	})
}
