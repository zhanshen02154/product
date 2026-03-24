package subscriber

import (
	"context"
	"fmt"
	event2 "github.com/zhanshen02154/product/internal/infrastructure/event"
	"go-micro.dev/v4/server"

	"github.com/zhanshen02154/product/internal/application/service"
	"github.com/zhanshen02154/product/internal/domain/event"
	"github.com/zhanshen02154/product/internal/domain/event/order"
	"go-micro.dev/v4/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 支付相关事件类型常量
const (
	EventTypeOnPaymentSuccess = "OnPaymentSuccess"
	TopicOrderEvents          = "OrderEvent"
	ConsumerGroupProduct      = "product-consumer"
)

// PaymentEventHandler 支付事件处理器接口（保持向后兼容）
type PaymentEventHandler interface {
	OnPaymentSuccess(ctx context.Context, req *order.OnPaymentSuccess) error
}

// paymentEventHandlerImpl 支付事件处理器实现类
type paymentEventHandlerImpl struct {
	productAppService service.IProductApplicationService
}

// NewPaymentEventHandler 新建Handler
func NewPaymentEventHandler(appService service.IProductApplicationService) PaymentEventHandler {
	return &paymentEventHandlerImpl{productAppService: appService}
}

// OnPaymentSuccess 支付成功回调事件（保持原有方法不变）
func (h *paymentEventHandlerImpl) OnPaymentSuccess(ctx context.Context, req *order.OnPaymentSuccess) error {
	if req.OrderDetails == nil {
		return status.Error(codes.InvalidArgument, "inventory cannot be nil")
	}
	if req.OrderId == 0 || len(req.OrderDetails) == 0 {
		return status.Error(codes.InvalidArgument, "orderId or products cannot be empty")
	}
	return h.productAppService.DeductInventory(ctx, req)
}

// ============ 新增：支持 EventDispatcher 的适配器方法 ============

// AsEventHandlers 将 PaymentEventHandler 转换为 EventHandler 列表，用于注册到 EventDispatcher
func (h *paymentEventHandlerImpl) AsEventHandlers() []event2.EventHandler {
	return []event2.EventHandler{
		// 支付成功事件处理器
		event2.NewGenericHandler(
			EventTypeOnPaymentSuccess,
			h.OnPaymentSuccess, // 直接使用现有方法
			func() *order.OnPaymentSuccess { return &order.OnPaymentSuccess{} },
		),
	}
}

// RegisterToDispatcher 注册到事件分发器
func (h *paymentEventHandlerImpl) RegisterToDispatcher(dispatcher *event2.EventDispatcher) error {
	handlers := h.AsEventHandlers()
	for _, handler := range handlers {
		if err := dispatcher.RegisterHandler(handler, TopicOrderEvents, ConsumerGroupProduct); err != nil {
			return fmt.Errorf("failed to register handler %s: %w", handler.EventType(), err)
		}
		logger.Infof("registered payment event handler: %s", handler.EventType())
	}
	return nil
}

// ============ 保留：向后兼容的注册方法 ============

// RegisterSubscriber 注册订阅器（保留向后兼容）
// Deprecated: 建议使用 RegisterToDispatcher 注册到 EventDispatcher
func (h *paymentEventHandlerImpl) RegisterSubscriber(srv server.Server) {
	logger.Warn("RegisterSubscriber is deprecated, please use RegisterToDispatcher with EventDispatcher")
}

// ============ 可选：便捷的包装函数 ============

// WrapPaymentEventHandler 包装支付事件处理器，支持直接处理 BaseEvent
func WrapPaymentEventHandler(h PaymentEventHandler) event2.EventHandlerFunc {
	return func(ctx context.Context, baseEvent *event.BaseEvent) error {
		// 反序列化具体事件
		req := &order.OnPaymentSuccess{}
		if err := event2.UnmarshalPayload(baseEvent, req); err != nil {
			return fmt.Errorf("failed to unmarshal OnPaymentSuccess: %w", err)
		}

		// 调用原有方法
		return h.OnPaymentSuccess(ctx, req)
	}
}
