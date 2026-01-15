package event

import (
	"context"
	"github.com/zhanshen02154/product/internal/config"
	"go-micro.dev/v4/client"
)

// Listener 事件总线
type Listener interface {
	Publish(ctx context.Context, topic string, event interface{}, key string, opts ...client.PublishOption) error
	Register(topic string, c client.Client) bool
	UnRegister(topic string) bool
	Close()
	Start()
}

// RegisterPublisher 注册发布事件
func RegisterPublisher(conf *config.Broker, eb Listener, c client.Client) {
	if len(conf.Publisher) > 0 {
		for i := range conf.Publisher {
			eb.Register(conf.Publisher[i], c)
		}
	}
}
