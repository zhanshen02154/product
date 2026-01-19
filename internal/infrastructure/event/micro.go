package event

import (
	"context"
	"errors"
	"github.com/Shopify/sarama"
	"github.com/zhanshen02154/product/internal/infrastructure/event/monitor"
	"go-micro.dev/v4"
	"go-micro.dev/v4/broker"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/metadata"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	partitionKey   = "Pkey"
	traceparentKey = "Traceparent"
)

// 事件侦听器
// 异步发送事件
type microListener struct {
	mu             sync.RWMutex
	eventPublisher sync.Map
	successChan    chan *sarama.ProducerMessage
	errorChan      chan *sarama.ProducerError
	wg             sync.WaitGroup
	quitChan       chan struct{}
	// started 用于防止重复 Start
	started bool
	opts    *options
}

// Publish 发布
func (l *microListener) Publish(ctx context.Context, topic string, msg interface{}, key string, opts ...client.PublishOption) error {
	if pub, ok := l.eventPublisher.Load(topic); ok {
		if e, assertOk := pub.(micro.Event); assertOk {
			// 将key放到metadata
			if key != "" {
				if _, ok := metadata.Get(ctx, partitionKey); !ok {
					ctx = metadata.Set(ctx, partitionKey, key)
				}
			}
			return e.Publish(ctx, msg, opts...)
		} else {
			return errors.New("invalid event")
		}
	}
	return errors.New("event not found")
}

// Register 注册
func (l *microListener) Register(topic string, c client.Client) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.eventPublisher.Store(topic, micro.NewEvent(topic, c))
	logger.Info("event ", topic, " registered")
	return true
}

// UnRegister 取消注册
func (l *microListener) UnRegister(topic string) bool {
	l.eventPublisher.Delete(topic)
	logger.Info("event: ", topic, " unregistered")
	return true
}

// Close 关闭
func (l *microListener) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	// 如果已经关闭或未初始化，直接返回
	if l.quitChan == nil {
		return
	}
	l.started = false
	l.eventPublisher.Range(func(key, value any) bool {
		l.eventPublisher.Delete(key)
		return true
	})

	if l.quitChan != nil {
		close(l.quitChan)
	}

	// 在所有 handler goroutine 退出后，再把引用置为 nil，避免竞争条件
	if l.quitChan != nil {
		l.quitChan = nil
	}
}

// Start 启动
func (l *microListener) Start() {
	l.mu.Lock()
	if l.started {
		l.mu.Unlock()
		return
	}
	l.started = true
	l.mu.Unlock()

	l.watchKafkaPipeline()
}

// 监听管道
func (l *microListener) watchKafkaPipeline() {
	l.wg.Add(2)
	go l.handleSuccess()
	go l.handleErrors()

}

// handleSuccess 处理发布成功的逻辑
func (l *microListener) handleSuccess() {
	defer l.wg.Done()
	for {
		select {
		case success, ok := <-l.successChan:
			if !ok {
				return
			}
			l.handleCallback(success, nil)
		case <-l.quitChan:
			logger.Info("Successes handler received stop signal.")
			return
		}
	}
}

// handleErrors 处理发布失败的逻辑
func (l *microListener) handleErrors() {
	defer l.wg.Done()
	for {
		select {
		case errMsg, ok := <-l.errorChan:
			if !ok {
				return
			}
			if errMsg != nil {
				l.handleCallback(errMsg.Msg, errMsg.Err)
			}
		case <-l.quitChan:
			logger.Info("Errors handler received stop signal.")
			return
		}
	}
}

// 处理回调信息
func (l *microListener) handleCallback(sg *sarama.ProducerMessage, err error) {
	if sg == nil || sg.Metadata == nil {
		return
	}
	msg, ok := sg.Metadata.(*broker.Message)
	if !ok || msg == nil {
		return
	}
	if msg.Header == nil {
		return
	}
	if v, ok := msg.Header[traceparentKey]; ok {
		msg.Header[strings.ToLower(traceparentKey)] = v
	}
	ctx := metadata.NewContext(context.Background(), msg.Header)
	ctx = context.WithValue(ctx, partitionContextKey{}, sg.Partition)
	ctx = context.WithValue(ctx, offsetKey{}, sg.Offset)

	fn := func(ctx context.Context, msg *broker.Message, err error) {
		if topic, ok := msg.Header["Micro-Topic"]; ok {
			if err == nil {
				monitor.MessageProducedCount.WithLabelValues(topic, "success", l.opts.name, l.opts.version).Inc()
				monitor.MessagesInFlight.WithLabelValues(topic, l.opts.name, l.opts.version).Dec()
			} else {
				monitor.MessageProducedCount.WithLabelValues(topic, "failure", l.opts.name, l.opts.version).Inc()
			}
			if _, ok := msg.Header["Timestamp"]; ok {
				startTime, convErr := strconv.ParseInt(msg.Header["Timestamp"], 10, 64)
				if convErr == nil {
					duration := time.Now().Sub(time.UnixMilli(startTime)).Seconds() * 1e3
					monitor.ProduceDuration.WithLabelValues(topic, l.opts.name, l.opts.version).Observe(duration)
				}
			}
		}
	}
	for i := len(l.opts.wrappers); i > 0; i-- {
		fn = l.opts.wrappers[i-1](fn)
	}
	fn(ctx, msg, err)
}

// NewListener 新建侦听器
func NewListener(opts ...Option) Listener {
	listener := microListener{
		mu:             sync.RWMutex{},
		eventPublisher: sync.Map{},
		wg:             sync.WaitGroup{},
		quitChan:       make(chan struct{}),
		opts: &options{
			wrappers: make([]PublishCallbackWrapper, 0, 10),
		},
	}
	for _, opt := range opts {
		opt(&listener)
	}
	return &listener
}
