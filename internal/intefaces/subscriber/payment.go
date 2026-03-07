package subscriber

import (
	"context"
	"github.com/zhanshen02154/product/internal/application/service"
	"github.com/zhanshen02154/product/internal/domain/event/order"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PaymentEventHandler 支付事件处理器接口
type PaymentEventHandler interface {
	OnPaymentSuccess(ctx context.Context, req *order.OnPaymentSuccess) error
	RegisterSubscriber(srv server.Server)
}

// PaymentEventHandlerImpl 支付事件处理器实现类
type paymentEventHandlerImpl struct {
	productAppService service.IProductApplicationService
}

// NewPaymentEventHandler 新建Handler
func NewPaymentEventHandler(appService service.IProductApplicationService) PaymentEventHandler {
	return &paymentEventHandlerImpl{productAppService: appService}
}

// OnPaymentSuccess OnPaySuccess 支付成功回调事件
func (h *paymentEventHandlerImpl) OnPaymentSuccess(ctx context.Context, req *order.OnPaymentSuccess) error {
	if req.OrderDetails == nil {
		return status.Error(codes.InvalidArgument, "inventory cannot be nil")
	}
	if req.OrderId == 0 || len(req.OrderDetails) == 0 {
		return status.Error(codes.InvalidArgument, "orderId or products cannot be empty")
	}
	return h.productAppService.DeductInventory(ctx, req)
}

// RegisterSubscriber 注册订阅器
func (h *paymentEventHandlerImpl) RegisterSubscriber(srv server.Server) {
	var err error
	queue := server.SubscriberQueue("product-consumer")
	err = micro.RegisterSubscriber("OnPaymentSuccess", srv, h.OnPaymentSuccess, queue)
	if err != nil {
		logger.Errorf("failed to register subscriber, error: %s", err.Error())
	}
}
