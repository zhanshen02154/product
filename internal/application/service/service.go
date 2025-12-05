package service

import (
	"context"
	"fmt"
	"github.com/zhanshen02154/product/internal/application/dto"
	"github.com/zhanshen02154/product/internal/domain/event/product"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/service"
	"github.com/zhanshen02154/product/internal/infrastructure"
	"github.com/zhanshen02154/product/internal/infrastructure/event"
	"github.com/zhanshen02154/product/pkg/swap"
	"go-micro.dev/v4/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IProductApplicationService interface {
	AddProduct(ctx context.Context, productInfo *dto.ProductDto) (*dto.AddProductResponse, error)
	DeductInventory(ctx context.Context, req *dto.OrderProductInvetoryDto) error
	DeductInvetoryRevert(ctx context.Context, req *dto.OrderProductInvetoryDto) error
}

// ProductApplicationService 商品服务应用层
type ProductApplicationService struct {
	// 领域服务层
	productDomainService service.IProductDataService
	// 服务上下文
	serviceContext *infrastructure.ServiceContext
	// 事件总线
	eb event.Listener
}

func NewProductApplicationService(serviceContext *infrastructure.ServiceContext, eb event.Listener) IProductApplicationService {
	return &ProductApplicationService{
		productDomainService: service.NewProductDataService(serviceContext.NewProductRepository(), serviceContext.NewOrderInventoryEventRepo()),
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

// DeductInventory 扣减订单的库存
func (appService *ProductApplicationService) DeductInventory(ctx context.Context, req *dto.OrderProductInvetoryDto) error {
	eventExists, err := appService.productDomainService.FindEventExistsByOrderId(ctx, req.OrderId)
	if err != nil {
		return status.Errorf(codes.NotFound, "failed to check order inventory event on %d, error: %s", err.Error())
	}
	if eventExists {
		return nil
	}
	err = appService.serviceContext.TxManager.Execute(ctx, func(txCtx context.Context) error {
		err = appService.productDomainService.DeductInventory(txCtx, req)
		if err != nil {
			return status.Errorf(codes.NotFound, "failed to deduct inventory on order %d, error: %s", req.OrderId, err.Error())
		}

		// 发布扣减库存成功事件
		inventoryDeductSuccessEvent := &product.OnInventoryDeductSuccess{
			OrderId: req.OrderId,
		}
		err = appService.eb.Publish(txCtx, "OnInventoryDeductSuccess", inventoryDeductSuccessEvent, fmt.Sprintf("%d", req.OrderId))
		if err != nil {
			return status.Errorf(codes.Aborted, "failed to publish event on %d, error: %s", req.OrderId, err.Error())
		}
		return nil
	})
	return err
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
