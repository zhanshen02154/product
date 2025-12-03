package wrapper

import (
	"context"
	"github.com/google/uuid"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/metadata"
	"strconv"
	"time"
)

// 元数据包装器
type MetaDataWrapper struct {
	client.Client
	serviceName    string
	serviceVersion string
}

// 发布
func (w *MetaDataWrapper) Publish(ctx context.Context, msg client.Message, opts ...client.PublishOption) error {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		md = make(map[string]string)
	}
	md["event_id"] = uuid.New().String()
	md["event_type"] = msg.Topic()
	md["timestamp"] = strconv.FormatInt(time.Now().Unix(), 36)
	md["source"] = w.serviceName
	md["schema_version"] = w.serviceVersion
	ctx = metadata.NewContext(ctx, md)
	return w.Client.Publish(ctx, msg, opts...)
}

// 新建包装器
func NewMetaDataWrapper(serviceName, serviceVersion string) func(client.Client) client.Client {
	return func(c client.Client) client.Client {
		return &MetaDataWrapper{
			Client:         c,
			serviceName:    serviceName,
			serviceVersion: serviceVersion,
		}
	}
}
