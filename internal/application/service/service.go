package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/zhanshen02154/product/internal/application/dto"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/service"
	"github.com/zhanshen02154/product/internal/infrastructure"
	"github.com/zhanshen02154/product/pkg/swap"
)

type IProductApplicationService interface {
	AddProduct(ctx context.Context, productInfo *dto.ProductDto) (*dto.AddProductResponse, error)
	DeductInvetory(ctx context.Context, req *dto.OrderProductInvetoryDto) error
	DeductInvetoryRevert(ctx context.Context, req *dto.OrderProductInvetoryDto) error
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
func (appService *ProductApplicationService) DeductInvetory(ctx context.Context, req *dto.OrderProductInvetoryDto) error {
	lock, err := appService.serviceContext.LockManager.NewLock(ctx, fmt.Sprintf("deductinvetory:%d", req.OrderId), 30)
	if err != nil {
		return err
	}
	ok, err := lock.TryLock(ctx)
	defer lock.UnLock(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New(fmt.Sprintf("duplicate notify: %d", req.OrderId))
	}
	return appService.serviceContext.TxManager.ExecuteWithBarrier(ctx, func(txCtx context.Context) error {
		return appService.productDomainService.DeductOrderInvetory(txCtx, req)
	})
}

// DeductInvetoryRevert 扣减订单的库存补偿
func (appService *ProductApplicationService) DeductInvetoryRevert(ctx context.Context, req *dto.OrderProductInvetoryDto) error {
	lock, err := appService.serviceContext.LockManager.NewLock(ctx, fmt.Sprintf("deductinvetoryrevert:%d", req.OrderId), 30)
	if err != nil {
		return err
	}
	ok, err := lock.TryLock(ctx)
	defer lock.UnLock(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New(fmt.Sprintf("duplicate notify: %d", req.OrderId))
	}
	return appService.serviceContext.TxManager.ExecuteWithBarrier(ctx, func(txCtx context.Context) error {
		return appService.productDomainService.DeductOrderInvetoryRevert(txCtx, req)
	})
}
