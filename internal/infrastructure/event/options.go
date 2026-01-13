package event

import (
	"context"
	"github.com/Shopify/sarama"
	"go-micro.dev/v4/broker"
)

type options struct {
	wrappers []PublishCallbackWrapper
	name     string
	version  string
}

type PublishCallbackFunc func(ctx context.Context, msg *broker.Message, err error)

type PublishCallbackWrapper func(PublishCallbackFunc) PublishCallbackFunc

type partitionContextKey struct{}

type offsetKey struct{}

type Option func(listener *microListener)

// 获取分区
func getPartition(ctx context.Context) int32 {
	if val, ok := ctx.Value(partitionContextKey{}).(int32); ok {
		return val
	}
	return -1
}

// 获取偏移量
func getOffset(ctx context.Context) int64 {
	if val, ok := ctx.Value(offsetKey{}).(int64); ok {
		return val
	}
	return -1
}

// WrapPublishCallback 应用包装器
func WrapPublishCallback(opts ...PublishCallbackWrapper) Option {
	return func(listener *microListener) {
		var wrps []PublishCallbackWrapper
		for _, o := range opts {
			wrps = append(wrps, o)
			listener.opts.wrappers = wrps
		}
	}
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

// WithServiceName 名称
func WithServiceName(name string) Option {
	return func(l *microListener) {
		l.opts.name = name
	}
}

// WithServiceVersion 版本
func WithServiceVersion(version string) Option {
	return func(l *microListener) {
		l.opts.version = version
	}
}
