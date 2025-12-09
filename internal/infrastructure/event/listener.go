package event

import (
	"context"
	"github.com/zhanshen02154/product/internal/config"
	"go-micro.dev/v4/client"
)

// Listener 侦听器
type Listener interface {
	Publish(ctx context.Context, topic string, msg interface{}, key string, opts ...client.PublishOption) error
	Register(topic string) bool
	UnRegister(topic string) bool
	Close()
}

// RegisterPublisher 注册事件发布器
func RegisterPublisher(conf *config.Broker, eb Listener) bool {
	if eb == nil || conf.Publisher == nil {
		return false
	}
	if len(conf.Publisher) > 0 {
		for i := range conf.Publisher {
			eb.Register(conf.Publisher[i])
		}
	}
	return true
}
