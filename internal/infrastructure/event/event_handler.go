package event

import (
	"context"
	"fmt"

	"github.com/zhanshen02154/product/internal/domain/event"
	"google.golang.org/protobuf/proto"
)

// EventHandler 事件处理器接口
type EventHandler interface {
	// EventType 返回处理器关注的事件类型
	EventType() string
	// Handle 处理事件
	Handle(ctx context.Context, baseEvent *event.BaseEvent) error
}

// EventHandlerFunc 事件处理函数类型
type EventHandlerFunc func(ctx context.Context, baseEvent *event.BaseEvent) error

// UnmarshalPayload 反序列化 Payload 到目标消息
func UnmarshalPayload[T proto.Message](baseEvent *event.BaseEvent, target T) error {
	if baseEvent == nil {
		return fmt.Errorf("base event is nil")
	}
	if len(baseEvent.Payload) == 0 {
		return fmt.Errorf("payload is empty")
	}
	if err := proto.Unmarshal(baseEvent.Payload, target); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	return nil
}
