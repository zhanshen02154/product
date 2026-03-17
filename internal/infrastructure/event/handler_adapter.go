package event

import (
	"context"
	"fmt"

	"github.com/zhanshen02154/product/internal/domain/event"
	"google.golang.org/protobuf/proto"
)

// GenericHandlerFunc 通用事件处理函数类型
// T 是具体的事件类型（如 order.OnPaymentSuccess）
type GenericHandlerFunc[T proto.Message] func(ctx context.Context, event T) error

// genericEventHandler 通用事件处理器适配器
type genericEventHandler[T proto.Message] struct {
	eventType   string
	handlerFunc GenericHandlerFunc[T]
	newMessage  func() T // 创建新消息实例的工厂函数
}

// NewGenericHandler 创建通用事件处理器
// eventType: 事件类型（如 "order.OnPaymentSuccess"）
// handlerFunc: 处理函数
// newMessage: 创建具体事件消息的工厂函数（如 func() proto.Message { return &order.OnPaymentSuccess{} }）
func NewGenericHandler[T proto.Message](
	eventType string,
	handlerFunc GenericHandlerFunc[T],
	newMessage func() T,
) EventHandler {
	return &genericEventHandler[T]{
		eventType:   eventType,
		handlerFunc: handlerFunc,
		newMessage:  newMessage,
	}
}

// EventType 返回事件类型
func (h *genericEventHandler[T]) EventType() string {
	return h.eventType
}

// Handle 处理事件
func (h *genericEventHandler[T]) Handle(ctx context.Context, baseEvent *event.BaseEvent) error {
	// 反序列化 Payload
	msg := h.newMessage()
	if err := UnmarshalPayload(baseEvent, msg); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	// 调用具体的处理函数
	return h.handlerFunc(ctx, msg)
}

// ============ 便捷适配器创建函数 ============

// HandlerAdapter 处理器适配器，用于将旧的 Handler 接口适配到新的 EventHandler 接口
type HandlerAdapter struct {
	eventType string
	handler   EventHandlerFunc
}

// NewHandlerAdapter 创建处理器适配器
func NewHandlerAdapter(eventType string, handler EventHandlerFunc) EventHandler {
	return &HandlerAdapter{
		eventType: eventType,
		handler:   handler,
	}
}

// EventType 返回事件类型
func (a *HandlerAdapter) EventType() string {
	return a.eventType
}

// Handle 处理事件
func (a *HandlerAdapter) Handle(ctx context.Context, baseEvent *event.BaseEvent) error {
	return a.handler(ctx, baseEvent)
}
