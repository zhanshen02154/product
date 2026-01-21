package subscriber

import (
	"context"
	"github.com/zhanshen02154/product/internal/application/dto"
	"github.com/zhanshen02154/product/internal/application/service"
	"github.com/zhanshen02154/product/internal/domain/event/order"
	"github.com/zhanshen02154/product/internal/domain/event/product"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
)

// PaymentEventHandler 支付事件处理器接口
type PaymentEventHandler interface {
	OnPaymentSuccess(ctx context.Context, req *order.OnPaymentSuccess) error
	RegisterSubscriber(srv server.Server)
}

// PaymentEventHandlerImpl 支付事件处理器实现类
type paymentEventHandlerImpl struct {
	productAppService service.IProductApplicationService
	orderInvetoryPool sync.Pool
}

// NewPaymentEventHandler 新建Handler
func NewPaymentEventHandler(appService service.IProductApplicationService) PaymentEventHandler {
	return &paymentEventHandlerImpl{productAppService: appService, orderInvetoryPool: sync.Pool{
		New: func() interface{} {
			return &dto.OrderProductInvetoryDto{
				OrderId:             0,
				ProductInvetory:     make([]*dto.OrderProductInvetoryItem, 0),
				ProductSizeInvetory: make([]*dto.OrderProductSizeInvetoryItem, 0),
			}
		},
	}}
}

// OnPaymentSuccess OnPaySuccess 支付成功回调事件
func (h *paymentEventHandlerImpl) OnPaymentSuccess(ctx context.Context, req *order.OnPaymentSuccess) error {
	if req.Products == nil {
		return status.Error(codes.InvalidArgument, "inventory cannot be nil")
	}
	if req.OrderId == 0 || len(req.Products) == 0 {
		return status.Error(codes.InvalidArgument, "orderId or products cannot be empty")
	}
	inventoryDto := h.orderInvetoryPool.Get().(*dto.OrderProductInvetoryDto)
	defer func() {
		inventoryDto.Reset()
		h.orderInvetoryPool.Put(inventoryDto)
	}()
	inventoryDto.ConvertTo(req)

	return h.productAppService.DeductInventory(ctx, inventoryDto)
}

// OnInventoryDeductFailed 库存扣减死信队列
func (h *paymentEventHandlerImpl) OnInventoryDeductFailed(ctx context.Context, req *product.OnInventoryDeductSuccess) error {
	inventoryDto := h.orderInvetoryPool.Get().(*dto.OrderProductInvetoryDto)
	defer func() {
		inventoryDto.Reset()
		h.orderInvetoryPool.Put(inventoryDto)
	}()
	inventoryDto.OrderId = req.OrderId
	for _, item := range req.Products {
		inventoryDto.ProductInvetory = append(inventoryDto.ProductInvetory, &dto.OrderProductInvetoryItem{
			Id:    item.Id,
			Count: item.Count,
		})
	}
	for _, item := range req.ProductSizes {
		inventoryDto.ProductSizeInvetory = append(inventoryDto.ProductSizeInvetory, &dto.OrderProductSizeInvetoryItem{
			Id:    item.Id,
			Count: item.Count,
		})
	}

	return h.productAppService.DeductInvetoryRevert(ctx, inventoryDto)
}

// RegisterSubscriber 注册订阅器
func (h *paymentEventHandlerImpl) RegisterSubscriber(srv server.Server) {
	var err error
	queue := server.SubscriberQueue("product-consumer")
	err = micro.RegisterSubscriber("OnPaymentSuccess", srv, h.OnPaymentSuccess, queue)
	if err != nil {
		logger.Errorf("failed to register subscriber, error: %s", err.Error())
	}
	err = micro.RegisterSubscriber("OnInventoryDeductSuccessDLQ", srv, h.OnInventoryDeductFailed, queue)
	if err != nil {
		logger.Errorf("failed to register subscriber, error: %s", err.Error())
	}
}
