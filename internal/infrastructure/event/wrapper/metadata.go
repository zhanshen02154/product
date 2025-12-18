package wrapper

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	metadatahelper "github.com/zhanshen02154/product/pkg/metadata"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/metadata"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"time"
)

// MetaDataWrapper 元数据包装器
type MetaDataWrapper struct {
	client.Client
	serviceName      string
	serviceVersion   string
	logger           *zap.Logger
	loggerFieldsPool sync.Pool
}

// Publish 发布
func (w *MetaDataWrapper) Publish(ctx context.Context, msg client.Message, opts ...client.PublishOption) error {
	startTime := time.Now()
	md, ok := metadata.FromContext(ctx)
	if !ok {
		md = make(map[string]string)
	}
	eventId := uuid.New().String()
	md["event_id"] = eventId
	md["event_type"] = msg.Topic()
	md["timestamp"] = strconv.FormatInt(startTime.Unix(), 10)
	md["source"] = w.serviceName
	md["schema_version"] = w.serviceVersion
	ctx = metadata.NewContext(ctx, md)
	err := w.Client.Publish(ctx, msg, opts...)
	duration := time.Since(startTime).Milliseconds()

	logFields := w.loggerFieldsPool.Get().([]zap.Field)
	defer func() {
		logFields = make([]zap.Field, 0)
		w.loggerFieldsPool.Put(logFields)
	}()
	logFields = append(logFields,
		zap.String("type", "publish"),
		zap.String("trace_id", metadatahelper.GetValueFromMetadata(ctx, "Trace_id")),
		zap.String("event_id", metadatahelper.GetValueFromMetadata(ctx, eventId)),
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

// NewMetaDataWrapper 新建包装器
func NewMetaDataWrapper(serviceName, serviceVersion string, zapLogger *zap.Logger) func(client.Client) client.Client {
	return func(c client.Client) client.Client {
		return &MetaDataWrapper{
			Client:         c,
			serviceName:    serviceName,
			serviceVersion: serviceVersion,
			loggerFieldsPool: sync.Pool{New: func() interface{} {
				return make([]zap.Field, 0)
			}},
			logger: zapLogger,
		}
	}
}
