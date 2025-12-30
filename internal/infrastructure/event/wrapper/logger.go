package wrapper

import (
	"context"
	"fmt"
	metadatahelper "github.com/zhanshen02154/product/pkg/metadata"
	"go-micro.dev/v4/client"
	"go.uber.org/zap"
	"sync"
	"time"
)

// 日志记录
type logWrapper struct {
	client.Client
	loggerFieldsPool sync.Pool
	logger           *zap.Logger
}

// Publish 发布事件
func (w *logWrapper) Publish(ctx context.Context, msg client.Message, opts ...client.PublishOption) error {
	logFields := w.loggerFieldsPool.Get().([]zap.Field)
	defer func() {
		logFields = make([]zap.Field, 0)
		w.loggerFieldsPool.Put(logFields)
	}()
	startTime := time.Now()
	err := w.Client.Publish(ctx, msg, opts...)
	duration := time.Since(startTime).Milliseconds()
	logFields = append(logFields,
		zap.String("type", "publish"),
		zap.String("trace_id", metadatahelper.GetTraceIdFromSpan(ctx)),
		zap.String("event_id", metadatahelper.GetValueFromMetadata(ctx, "Event_id")),
		zap.String("topic", msg.Topic()),
		zap.String("source", metadatahelper.GetValueFromMetadata(ctx, "Source")),
		zap.String("schema_version", metadatahelper.GetValueFromMetadata(ctx, "Schema_version")),
		zap.Int64("published_at", startTime.Unix()),
		zap.String("remote", metadatahelper.GetValueFromMetadata(ctx, "Remote")),
		zap.String("accept_encoding", metadatahelper.GetValueFromMetadata(ctx, "Accept-Encoding")),
		zap.String("key", metadatahelper.GetValueFromMetadata(ctx, "Pkey")),
		zap.Int64("duration", duration),
	)

	if err != nil {
		w.logger.Error(fmt.Sprintf("failed to publish event %s, error: %s", msg.Topic(), err.Error()), logFields...)
	} else {
		w.logger.Info(fmt.Sprintf("publish event %s success", msg.Topic()), logFields...)
	}

	return err
}

// NewClientLogWrapper 新建客户端日志包装器
func NewClientLogWrapper(zapLogger *zap.Logger) func(client.Client) client.Client {
	return func(c client.Client) client.Client {
		return &logWrapper{
			Client: c,
			loggerFieldsPool: sync.Pool{New: func() interface{} {
				return make([]zap.Field, 0)
			}},
			logger: zapLogger,
		}
	}
}
