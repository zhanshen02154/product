package event

import (
	"context"
	"fmt"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/metadata"
	"sync"
)

const partitionKey = "Pkey"

// microListener 侦听器
type microListener struct {
	mu             sync.RWMutex
	eventPublisher map[string]micro.Event
	c              client.Client
}

// Publish 发布
func (l *microListener) Publish(ctx context.Context, topic string, msg interface{}, key string, opts ...client.PublishOption) error {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if _, ok := l.eventPublisher[topic]; !ok {
		return fmt.Errorf("topic: %s event not registerd", topic)
	}

	// 将key放到metadata
	if key != "" {
		if _, ok := metadata.Get(ctx, partitionKey); !ok {
			ctx = metadata.Set(ctx, partitionKey, key)
		}
	}
	return l.eventPublisher[topic].Publish(ctx, msg, opts...)
}

// Register 注册
func (l *microListener) Register(topic string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.eventPublisher[topic]; !ok {
		l.eventPublisher[topic] = micro.NewEvent(topic, l.c)
	}
	logger.Info("event ", topic, " was registered")
	return true
}

// UnRegister 取消注册
func (l *microListener) UnRegister(topic string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.eventPublisher == nil {
		return true
	}
	if len(l.eventPublisher) == 0 {
		return true
	}
	if _, ok := l.eventPublisher[topic]; ok {
		delete(l.eventPublisher, topic)
		logger.Info("event: ", topic, " unregistered")
	}
	l.eventPublisher = nil
	return true
}

// Close 关闭
func (l *microListener) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.eventPublisher == nil {
		return
	}
	if len(l.eventPublisher) > 0 {
		return
	}
	for k, _ := range l.eventPublisher {
		delete(l.eventPublisher, k)
		logger.Info("event: ", k, " unregistered")
	}
}

// NewListener 新建侦听器
func NewListener(c client.Client) Listener {
	return &microListener{
		mu:             sync.RWMutex{},
		eventPublisher: make(map[string]micro.Event),
		c:              c,
	}
}
