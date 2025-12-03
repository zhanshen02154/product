package event

import (
	"context"
	"fmt"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/logger"
	"sync"
)

// MicroListener 侦听器
type MicroListener struct {
	mu             sync.RWMutex
	eventPublisher map[string]micro.Event
	c              client.Client
}

// Publish 发布
func (l *MicroListener) Publish(ctx context.Context, topic string, msg interface{}, opts ...client.PublishOption) error {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if _, ok := l.eventPublisher[topic]; !ok {
		return fmt.Errorf("topic: %s event not registerd", topic)
	}
	return l.eventPublisher[topic].Publish(ctx, msg, opts...)
}

// Register 注册
func (l *MicroListener) Register(topic string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.eventPublisher[topic]; !ok {
		l.eventPublisher[topic] = micro.NewEvent(topic, l.c)
	}
	logger.Info("event ", topic, " was registered")
	return true
}

// UnRegister 取消注册
func (l *MicroListener) UnRegister(topic string) bool {
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
	return true
}

// Close 关闭
func (l *MicroListener) Close() {
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
	return &MicroListener{
		mu:             sync.RWMutex{},
		eventPublisher: make(map[string]micro.Event),
		c:              c,
	}
}
