package wrapper

import (
	"context"
	"github.com/google/uuid"
	metadata2 "github.com/zhanshen02154/product/pkg/metadata"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/metadata"
	"strconv"
	"time"
)

// MetaDataWrapper 元数据包装器
type MetaDataWrapper struct {
	client.Client
	serviceName    string
	serviceVersion string
}

// Publish 发布
func (w *MetaDataWrapper) Publish(ctx context.Context, msg client.Message, opts ...client.PublishOption) error {
	startTime := time.Now()
	md, ok := metadata.FromContext(ctx)
	if !ok {
		md = make(map[string]string)
	}
	md["Trace_id"] = metadata2.GetTraceIdFromSpan(ctx)
	md["Event_id"] = uuid.New().String()
	md["Timestamp"] = strconv.FormatInt(startTime.Unix(), 10)
	md["Source"] = w.serviceName
	md["Schema_version"] = w.serviceVersion
	ctx = metadata.NewContext(ctx, md)
	return w.Client.Publish(ctx, msg, opts...)
}

// NewMetaDataWrapper 新建包装器
func NewMetaDataWrapper(serviceName, serviceVersion string) func(client.Client) client.Client {
	return func(c client.Client) client.Client {
		return &MetaDataWrapper{
			Client:         c,
			serviceName:    serviceName,
			serviceVersion: serviceVersion,
		}
	}
}
