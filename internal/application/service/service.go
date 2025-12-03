package service

import (
	"context"
	"fmt"
	"github.com/zhanshen02154/product/internal/application/dto"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/service"
	"github.com/zhanshen02154/product/internal/infrastructure"
	"github.com/zhanshen02154/product/internal/infrastructure/event"
	"github.com/zhanshen02154/product/pkg/swap"
	"go-micro.dev/v4/logger"
)

type IProductApplicationService interface {
	AddProduct(ctx context.Context, productInfo *dto.ProductDto) (*dto.AddProductResponse, error)
	DeductInvetory(ctx context.Context, req *dto.OrderProductInvetoryDto) error
	DeductInvetoryRevert(ctx context.Context, req *dto.OrderProductInvetoryDto) error
}

// ProductApplicationService 商品服务应用层
type ProductApplicationService struct {
	// 领域服务层
	productDomainService service.IProductDataService
	// 服务上下文
	serviceContext *infrastructure.ServiceContext
	// 事件总线
	eb event.Bus
}

func NewProductApplicationService(serviceContext *infrastructure.ServiceContext, eb event.Bus) IProductApplicationService {
	return &ProductApplicationService{
		productDomainService: service.NewProductDataService(serviceContext.OrderRepository),
		serviceContext:       serviceContext,
		eb:                   eb,
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
	lock, err := appService.serviceContext.LockManager.NewLock(ctx, fmt.Sprintf("deductinvetory-%d", req.OrderId), 15)
	if err != nil {
		return err
	}
	if err := lock.TryLock(ctx); err != nil {
		return err
	}

	defer func() {
		if err := lock.UnLock(ctx); err != nil {
			logger.Error("failed to unlock: ", lock.GetKey(ctx), " reason: ", err)
		}
	}()
	return appService.serviceContext.TxManager.ExecuteWithBarrier(ctx, func(txCtx context.Context) error {
		return appService.productDomainService.DeductOrderInvetory(txCtx, req)
	})
}

// DeductInvetoryRevert 扣减订单的库存补偿
func (appService *ProductApplicationService) DeductInvetoryRevert(ctx context.Context, req *dto.OrderProductInvetoryDto) error {
	lock, err := appService.serviceContext.LockManager.NewLock(ctx, fmt.Sprintf("deductinvetoryrevert-%d", req.OrderId), 15)
	if err != nil {
		return err
	}
	if err := lock.TryLock(ctx); err != nil {
		return err
	}

	defer func() {
		if err := lock.UnLock(ctx); err != nil {
			logger.Error("failed to unlock: ", lock.GetKey(ctx), " reason: ", err)
		}
	}()
	return appService.serviceContext.TxManager.ExecuteWithBarrier(ctx, func(txCtx context.Context) error {
		return appService.productDomainService.DeductOrderInvetoryRevert(txCtx, req)
	})
}
