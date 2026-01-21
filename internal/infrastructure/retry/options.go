package retry

import (
	"github.com/zhanshen02154/product/internal/config"
	"go.uber.org/zap"
	"time"
)

type options struct {
	maxRetries      uint64
	initialInterval time.Duration
	maxInterval     time.Duration
	maxElapsedTime  time.Duration
	logger          *zap.Logger
}

type Option func(options *options)

// WithKafkaConsumerConfig 引用Kafka消费者重试配置
func WithKafkaConsumerConfig(conf *config.KafkaConsumer) Option {
	return func(o *options) {
		o.maxRetries = conf.Retry.MaxRetries
		o.initialInterval = time.Duration(conf.Retry.InitialInterval) * time.Second
		o.maxInterval = time.Duration(conf.Retry.MaxInterval) * time.Second
		o.maxElapsedTime = time.Duration(conf.Retry.MaxElapsedTime) * time.Second
	}
}

// WithLogger 日志
func WithLogger(logger *zap.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}
