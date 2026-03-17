package event

import (
	"context"
	"fmt"
	"sync"

	"github.com/zhanshen02154/product/internal/domain/event"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"
)

// EventDispatcher 事件分发器
type EventDispatcher struct {
	mu          sync.RWMutex
	handlers    map[string]EventHandler            // 事件类型 -> 处理器映射
	topicConfig map[string][]string                // topic -> 支持的事件类型列表
	topicMap    map[string]string                  // 事件类型 -> topic 映射
	consumerMap map[string]server.SubscriberOption // topic -> 消费者组配置
}

// NewEventDispatcher 创建事件分发器
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		handlers:    make(map[string]EventHandler),
		topicConfig: make(map[string][]string),
		topicMap:    make(map[string]string),
		consumerMap: make(map[string]server.SubscriberOption),
	}
}

// RegisterHandler 注册事件处理器
func (d *EventDispatcher) RegisterHandler(handler EventHandler, topic string, consumerGroup string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	eventType := handler.EventType()
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}

	// 注册处理器
	if _, exists := d.handlers[eventType]; exists {
		logger.Warnf("event handler for type '%s' already exists, will be overwritten", eventType)
	}
	d.handlers[eventType] = handler

	// 更新 topic 配置
	d.topicConfig[topic] = append(d.topicConfig[topic], eventType)
	d.topicMap[eventType] = topic

	// 设置消费者组
	if consumerGroup != "" {
		d.consumerMap[topic] = server.SubscriberQueue(consumerGroup)
	}

	logger.Infof("registered event handler: type=%s, topic=%s, consumerGroup=%s",
		eventType, topic, consumerGroup)
	return nil
}

// RegisterHandlers 批量注册事件处理器
func (d *EventDispatcher) RegisterHandlers(handlers []EventHandler, topic string, consumerGroup string) error {
	for _, handler := range handlers {
		if err := d.RegisterHandler(handler, topic, consumerGroup); err != nil {
			return err
		}
	}
	return nil
}

// Dispatch 分发事件到对应的处理器
func (d *EventDispatcher) Dispatch(ctx context.Context, baseEvent *event.BaseEvent) error {
	if baseEvent == nil {
		return fmt.Errorf("base event is nil")
	}

	d.mu.RLock()
	handler, exists := d.handlers[baseEvent.EventType]
	d.mu.RUnlock()

	if !exists {
		logger.Warnf("no handler found for event type: %s",
			baseEvent.EventType)
		return fmt.Errorf("no handler found for event type: %s", baseEvent.EventType)
	}

	logger.Infof("dispatching event: eventType=%s",
		baseEvent.EventType)

	return handler.Handle(ctx, baseEvent)
}

// RegisterSubscribers 注册所有订阅器
func (d *EventDispatcher) RegisterSubscribers(srv server.Server) error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// 为每个 topic 注册一个统一的订阅器
	for topic, eventTypes := range d.topicConfig {
		// 创建该 topic 的分发处理函数
		dispatchFunc := d.createDispatchFunc(topic)

		// 获取消费者组配置
		var opts []server.SubscriberOption
		if consumerOpt, exists := d.consumerMap[topic]; exists {
			opts = append(opts, consumerOpt)
		}

		// 注册订阅器
		err := micro.RegisterSubscriber(topic, srv, dispatchFunc, opts...)
		if err != nil {
			return fmt.Errorf("failed to register subscriber for topic %s: %w", topic, err)
		}

		logger.Infof("registered subscriber for topic: %s, event types: %v", topic, eventTypes)
	}

	return nil
}

// createDispatchFunc 创建特定 topic 的分发函数
func (d *EventDispatcher) createDispatchFunc(topic string) func(ctx context.Context, baseEvent *event.BaseEvent) error {
	return func(ctx context.Context, baseEvent *event.BaseEvent) error {
		// 验证事件类型是否属于该 topic
		d.mu.RLock()
		_, belongsToTopic := d.topicMap[baseEvent.EventType]
		d.mu.RUnlock()

		if !belongsToTopic {
			logger.Warnf("event type %s does not belong to topic %s", baseEvent.EventType, topic)
			return fmt.Errorf("event type %s does not belong to topic %s", baseEvent.EventType, topic)
		}

		return d.Dispatch(ctx, baseEvent)
	}
}

// GetSupportedEventTypes 获取支持的事件类型列表
func (d *EventDispatcher) GetSupportedEventTypes() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	types := make([]string, 0, len(d.handlers))
	for eventType := range d.handlers {
		types = append(types, eventType)
	}
	return types
}

// GetHandler 获取指定事件类型的处理器
func (d *EventDispatcher) GetHandler(eventType string) (EventHandler, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	handler, exists := d.handlers[eventType]
	return handler, exists
}
