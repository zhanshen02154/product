package event

import (
	"context"
	"go-micro.dev/v4/client"
)

// 事件总线
type Bus interface {
	Publish(ctx context.Context, topic string, event interface{}, opts ...client.PublishOption) error
	Register(topic string) bool
	UnRegister(topic string) bool
	Close()
}
