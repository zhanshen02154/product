package event

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"go-micro.dev/v4"
	"go-micro.dev/v4/broker"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/metadata"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	partitionKey     = "Pkey"
	traceparentKey   = "Traceparent"
	deadletterSuffix = "DLQ"
)

// 事件侦听器
// 异步发送事件
type microListener struct {
	mu                   sync.RWMutex
	eventPublisher       map[string]micro.Event
	c                    client.Client
	logger               *zap.Logger
	b                    broker.Broker
	successChan          chan *sarama.ProducerMessage
	errorChan            chan *sarama.ProducerError
	wg                   sync.WaitGroup
	publishTimeThreshold int64
	quitChan             chan struct{}
	// started 用于防止重复 Start
	started bool
}

type Option func(listener *microListener)

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
	err := l.eventPublisher[topic].Publish(ctx, msg, opts...)
	return err
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
	return true
}

// Close 关闭
func (l *microListener) Close() {
	l.mu.Lock()
	// 如果已经关闭或未初始化，直接返回
	if l.eventPublisher == nil && l.quitChan == nil {
		l.mu.Unlock()
		return
	}
	if l.eventPublisher != nil {
		for k := range l.eventPublisher {
			delete(l.eventPublisher, k)
			logger.Info("event: ", k, " unregistered")
		}
		l.eventPublisher = nil
	}
	if l.quitChan != nil {
		close(l.quitChan)
		l.quitChan = nil
	}
	l.mu.Unlock()

	l.wg.Wait()
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

func (l *microListener) watchKafkaPipeline() {
	propagator := otel.GetTextMapPropagator()
	tracer := otel.Tracer("kafka-producer-internal")
	l.wg.Add(2)
	go l.handleSuccess(propagator, tracer)
	go l.handleErrors(propagator, tracer)

}

// handleSuccess 处理发布成功的逻辑
func (l *microListener) handleSuccess(propagator propagation.TextMapPropagator, tracer trace.Tracer) {
	defer l.wg.Done()
	for {
		select {
		case success, ok := <-l.successChan:
			if !ok {
				return
			}
			l.handleCallback(success, nil, propagator, tracer)
		case <-l.quitChan:
			logger.Info("Successes handler received stop signal.")
			return
		}
	}
}

// handleErrors 处理发布失败的逻辑
func (l *microListener) handleErrors(propagator propagation.TextMapPropagator, tracer trace.Tracer) {
	defer l.wg.Done()
	for {
		select {
		case errMsg, ok := <-l.errorChan:
			if !ok {
				return
			}
			if errMsg != nil {
				l.handleCallback(errMsg.Msg, errMsg.Err, propagator, tracer)
			}
		case <-l.quitChan:
			logger.Info("Errors handler received stop signal.")
			return
		}
	}
}

// 处理回调信息
func (l *microListener) handleCallback(sg *sarama.ProducerMessage, err error, propagator propagation.TextMapPropagator, tracer trace.Tracer) {
	if sg == nil || sg.Metadata == nil {
		return
	}
	msg, ok := sg.Metadata.(*broker.Message)
	if !ok || msg == nil {
		return
	}
	if msg.Header == nil {
		msg.Header = make(map[string]string)
	}
	if v, ok := msg.Header[traceparentKey]; ok {
		msg.Header[strings.ToLower(traceparentKey)] = v
	}

	// 从header提取数据
	parentCtx := propagator.Extract(context.Background(), propagation.MapCarrier(msg.Header))
	spanName := "Kafka Async Callback"
	_, span := tracer.Start(parentCtx, spanName, trace.WithSpanKind(trace.SpanKindProducer))
	defer span.End()

	if err != nil {
		// 不是死信队列则投递到死信队列里（复制 header 以避免并发修改）
		if !strings.HasSuffix(sg.Topic, deadletterSuffix) && l.b != nil {
			dlMsg := &broker.Message{
				Header: make(map[string]string, len(msg.Header)+2),
				Body:   msg.Body,
			}
			for k, v := range msg.Header {
				dlMsg.Header[k] = v
			}
			dlMsg.Header["Micro-Topic"] = sg.Topic + deadletterSuffix
			dlMsg.Header["error"] = err.Error()
			if perr := l.b.Publish(dlMsg.Header["Micro-Topic"], dlMsg); perr != nil {
				logger.Errorf("failed to publish dead letter topic %s on id %s, error: %s", dlMsg.Header["Micro-Topic"], dlMsg.Header["Event_id"], perr.Error())
			}
		}
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	} else {
		span.SetAttributes(
			attribute.Int64("kafka.offset", sg.Offset),
			attribute.Int("kafka.partition", int(sg.Partition)),
			attribute.String("kafka.topic", sg.Topic),
		)
	}

	l.logPublish(msg, err)
}

// 记录发布成功日志
func (l *microListener) logPublish(msg *broker.Message, msgErr error) {
	var logFields []zap.Field
	if msg == nil {
		return
	}
	var duration int64
	currentTime := time.Now()
	if _, ok := msg.Header["Timestamp"]; ok {
		startTimeInt, err := strconv.ParseInt(msg.Header["Timestamp"], 10, 64)
		if err != nil {
			logger.Errorf("failed to convert publush event timestamp: %s in topic: %s", err.Error())
			return
		}
		startTime := time.UnixMilli(startTimeInt)
		duration = currentTime.Sub(startTime).Milliseconds()
	} else {
		duration = -1
	}
	pKey := ""
	if _, ok := msg.Header[partitionKey]; ok {
		pKey = msg.Header[partitionKey]
	}
	logFields = append(logFields,
		zap.String("type", "publish"),
		zap.String("trace_id", msg.Header["Trace_id"]),
		zap.String("topic", msg.Header["Micro-Topic"]),
		zap.String("event_id", msg.Header["Event_id"]),
		zap.String("source", msg.Header["Source"]),
		zap.String("schema_version", msg.Header["Schema_version"]),
		zap.Int64("published_at", currentTime.Unix()),
		zap.String("remote", msg.Header["Remote"]),
		zap.String("accept_encoding", msg.Header["Accept-Encoding"]),
		zap.String("key", pKey),
		zap.Int64("duration", duration),
	)
	if msgErr != nil {
		l.logger.Error(fmt.Sprintf("failed to publish event to topic %s on event_id %s. error: %s", msg.Header["Micro-Topic"], msg.Header["Event_id"], msgErr.Error()), logFields...)
		return
	}
	if duration > l.publishTimeThreshold {
		l.logger.Warn(fmt.Sprintf("publish event %s too slow, greater than %d", msg.Header["Micro-Topic"], l.publishTimeThreshold), logFields...)
		return
	}
	l.logger.Info(fmt.Sprintf("published to topic %s success", msg.Header["Micro-Topic"]), logFields...)
}

// NewListener 新建侦听器
func NewListener(opts ...Option) Listener {
	listener := microListener{
		mu:             sync.RWMutex{},
		eventPublisher: make(map[string]micro.Event),
		wg:             sync.WaitGroup{},
		quitChan:       make(chan struct{}),
	}
	for _, opt := range opts {
		opt(&listener)
	}
	return &listener
}

// Successes 返回内部 success 通道（用于将 producer.Successes() 转发至此）
func (l *microListener) Successes() chan *sarama.ProducerMessage {
	return l.successChan
}

// Errors 返回内部 error 通道（用于将 producer.Errors() 转发至此）
func (l *microListener) Errors() chan *sarama.ProducerError {
	return l.errorChan
}

// WithProducerChannels 允许注入外部的 producer 通道
func WithProducerChannels(success chan *sarama.ProducerMessage, errc chan *sarama.ProducerError) Option {
	return func(l *microListener) {
		if success != nil {
			l.successChan = success
		}
		if errc != nil {
			l.errorChan = errc
		}
	}
}

// WithLogger 设置Logger
func WithLogger(l *zap.Logger) Option {
	return func(listener *microListener) {
		listener.logger = l
	}
}

// WithBroker 设置Broker
func WithBroker(b broker.Broker) Option {
	return func(l *microListener) {
		l.b = b
	}
}

// WithClient 设置客户端
func WithClient(c client.Client) Option {
	return func(l *microListener) {
		l.c = c
	}
}

// WithPulishTimeThreshold 发布时间
func WithPulishTimeThreshold(timeout int64) Option {
	return func(listener *microListener) {
		listener.publishTimeThreshold = timeout
	}
}
