package event

import (
	"context"
	"github.com/zhanshen02154/product/internal/intefaces/subscriber"
	"testing"

	"github.com/zhanshen02154/product/internal/domain/event"
	"github.com/zhanshen02154/product/internal/domain/event/order"
	"google.golang.org/protobuf/proto"
)

// TestEventDispatcher_RegisterHandler 测试注册处理器
func TestEventDispatcher_RegisterHandler(t *testing.T) {
	dispatcher := NewEventDispatcher()

	handler := NewHandlerAdapter("test.event", func(ctx context.Context, e *event.BaseEvent) error {
		return nil
	})

	// 测试注册
	err := dispatcher.RegisterHandler(handler, "test_topic", "test-consumer")
	if err != nil {
		t.Fatalf("failed to register handler: %v", err)
	}

	// 验证处理器已注册
	registeredHandler, exists := dispatcher.GetHandler("test.event")
	if !exists {
		t.Fatal("handler not found after registration")
	}

	if registeredHandler.EventType() != "test.event" {
		t.Errorf("expected event type 'test.event', got '%s'", registeredHandler.EventType())
	}

	// 测试支持的事件类型列表
	eventTypes := dispatcher.GetSupportedEventTypes()
	if len(eventTypes) != 1 {
		t.Errorf("expected 1 event type, got %d", len(eventTypes))
	}
}

// TestEventDispatcher_Dispatch 测试事件分发
func TestEventDispatcher_Dispatch(t *testing.T) {
	dispatcher := NewEventDispatcher()

	processed := false
	handler := NewHandlerAdapter("test.event", func(ctx context.Context, e *event.BaseEvent) error {
		processed = true
		return nil
	})

	dispatcher.RegisterHandler(handler, "test_topic", "test-consumer")

	// 测试分发
	baseEvent := &event.BaseEvent{
		EventType: "test.event",
		Timestamp: 1234567890,
		Payload:   []byte{},
	}

	err := dispatcher.Dispatch(context.Background(), baseEvent)
	if err != nil {
		t.Fatalf("failed to dispatch event: %v", err)
	}

	if !processed {
		t.Error("handler was not processed")
	}
}

// TestEventDispatcher_Dispatch_UnknownEventType 测试未知事件类型
func TestEventDispatcher_Dispatch_UnknownEventType(t *testing.T) {
	dispatcher := NewEventDispatcher()

	baseEvent := &event.BaseEvent{
		EventType: "unknown.event",
	}

	err := dispatcher.Dispatch(context.Background(), baseEvent)
	if err == nil {
		t.Fatal("expected error for unknown event type")
	}
}

// TestGenericHandler 测试泛型处理器
func TestGenericHandler(t *testing.T) {
	// 创建测试事件
	testEvent := &order.OnPaymentSuccess{
		OrderId: 12345,
		OrderDetails: []*order.OrderDetail{
			{
				ProductId: 1,
				SkuId:     100,
				Quantity:  2,
			},
		},
	}

	// 序列化
	payload, err := proto.Marshal(testEvent)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	// 创建处理器
	processed := false
	handler := NewGenericHandler(
		"order.OnPaymentSuccess",
		func(ctx context.Context, e *order.OnPaymentSuccess) error {
			processed = true
			if e.OrderId != 12345 {
				t.Errorf("expected orderId 12345, got %d", e.OrderId)
			}
			return nil
		},
		func() *order.OnPaymentSuccess { return &order.OnPaymentSuccess{} },
	)

	// 测试 EventType
	if handler.EventType() != "order.OnPaymentSuccess" {
		t.Errorf("expected event type 'order.OnPaymentSuccess', got '%s'", handler.EventType())
	}

	// 测试 Handle
	baseEvent := &event.BaseEvent{
		EventType: "order.OnPaymentSuccess",
		Payload:   payload,
	}

	err = handler.Handle(context.Background(), baseEvent)
	if err != nil {
		t.Fatalf("failed to handle event: %v", err)
	}

	if !processed {
		t.Error("handler was not processed")
	}
}

// TestUnmarshalPayload 测试 Payload 反序列化
func TestUnmarshalPayload(t *testing.T) {
	testEvent := &order.OnPaymentSuccess{
		OrderId: 12345,
	}

	payload, err := proto.Marshal(testEvent)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// 测试成功反序列化
	target := &order.OnPaymentSuccess{}
	err = UnmarshalPayload(&event.BaseEvent{Payload: payload}, target)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if target.OrderId != 12345 {
		t.Errorf("expected orderId 12345, got %d", target.OrderId)
	}

	// 测试 nil BaseEvent
	err = UnmarshalPayload(nil, target)
	if err == nil {
		t.Fatal("expected error for nil base event")
	}

	// 测试空 Payload
	err = UnmarshalPayload(&event.BaseEvent{Payload: []byte{}}, target)
	if err == nil {
		t.Fatal("expected error for empty payload")
	}
}

// TestPaymentEventHandler_AsEventHandlers 测试 PaymentEventHandler 的适配方法
func TestPaymentEventHandler_AsEventHandlers(t *testing.T) {
	// 创建 mock service（实际项目中应使用 mock 框架）
	handler := subscriber.NewPaymentEventHandler(nil)

	// 类型断言获取 AsEventHandlers 方法
	adapter, ok := interface{}(handler).(interface {
		AsEventHandlers() []EventHandler
	})
	if !ok {
		t.Fatal("PaymentEventHandler does not implement AsEventHandlers")
	}

	// 获取 EventHandler 列表
	handlers := adapter.AsEventHandlers()

	if len(handlers) != 1 {
		t.Errorf("expected 1 handler, got %d", len(handlers))
	}

	if handlers[0].EventType() != subscriber.EventTypeOnPaymentSuccess {
		t.Errorf("expected event type '%s', got '%s'", subscriber.EventTypeOnPaymentSuccess, handlers[0].EventType())
	}
}

// TestEventDispatcher_RegisterHandlers 测试批量注册
func TestEventDispatcher_RegisterHandlers(t *testing.T) {
	dispatcher := NewEventDispatcher()

	handlers := []EventHandler{
		NewHandlerAdapter("event.1", func(ctx context.Context, e *event.BaseEvent) error { return nil }),
		NewHandlerAdapter("event.2", func(ctx context.Context, e *event.BaseEvent) error { return nil }),
	}

	err := dispatcher.RegisterHandlers(handlers, "test_topic", "test-consumer")
	if err != nil {
		t.Fatalf("failed to register handlers: %v", err)
	}

	eventTypes := dispatcher.GetSupportedEventTypes()
	if len(eventTypes) != 2 {
		t.Errorf("expected 2 event types, got %d", len(eventTypes))
	}
}
