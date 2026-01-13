package event

import (
	"context"
	"github.com/go-micro/plugins/v4/wrapper/trace/opentelemetry"
	"github.com/zhanshen02154/product/internal/config"
	"github.com/zhanshen02154/product/internal/infrastructure/event/monitor"
	"go-micro.dev/v4/broker"
	"go-micro.dev/v4/logger"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"strconv"
	"strings"
	"time"
)

const (
	deadletterSuffix = "DLQ"
)

type deadletterOptions struct {
	asyncBroker broker.Broker
	opts        struct {
		traceProvider trace.TracerProvider
		service       string
		version       string
	}
}

type DeadLetterOption func(*deadletterOptions)

func NewDeadletterWrapper(opts ...DeadLetterOption) PublishCallbackWrapper {
	dlqOptions := deadletterOptions{}
	for _, o := range opts {
		o(&dlqOptions)
	}
	return func(next PublishCallbackFunc) PublishCallbackFunc {
		return func(ctx context.Context, msg *broker.Message, err error) {
			// 先执行回调函数
			next(ctx, msg, err)
			if err == nil {
				return
			}
			topic := msg.Header["Micro-Topic"]
			if strings.HasSuffix(topic, deadletterSuffix) {
				return
			}
			spanOpts := []trace.SpanStartOption{
				trace.WithSpanKind(trace.SpanKindProducer),
			}
			topic = topic + deadletterSuffix
			newCtx, span := opentelemetry.StartSpanFromContext(ctx, dlqOptions.opts.traceProvider, "Pub to dead letter topic "+topic, spanOpts...)
			defer span.End()
			header := make(map[string]string)
			header["x-origin-topic"] = msg.Header["Micro-Topic"]
			header["x-error"] = err.Error()
			header["x-origin-timestamp"] = msg.Header["Timestamp"]
			for k, v := range msg.Header {
				if k == "Timestamp" || k == traceparentKey {
					continue
				}
				header[k] = v
			}
			header["Timestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
			header["Micro-Topic"] = topic
			header["Source"] = dlqOptions.opts.service
			header["Schema_version"] = dlqOptions.opts.version
			dlMsg := broker.Message{
				Header: header,
				Body:   msg.Body,
			}
			if pErr := dlqOptions.asyncBroker.Publish(topic, &dlMsg, broker.PublishContext(newCtx)); pErr != nil {
				logger.Error("Failed to publish dead letter topic " + topic + " error: " + pErr.Error())
				span.SetStatus(codes.Error, pErr.Error())
				span.RecordError(pErr)
			} else {
				monitor.MessagesInFlight.WithLabelValues(topic, dlqOptions.opts.service, dlqOptions.opts.version).Inc()
			}
			return
		}
	}
}

func WithBroker(b broker.Broker) DeadLetterOption {
	return func(d *deadletterOptions) {
		d.asyncBroker = b
	}
}

func WithTracer(tracerProvider trace.TracerProvider) DeadLetterOption {
	return func(d *deadletterOptions) {
		d.opts.traceProvider = tracerProvider
	}
}

// WithServiceInfo 引入服务信息
func WithServiceInfo(info *config.ServiceInfo) DeadLetterOption {
	return func(o *deadletterOptions) {
		o.opts.service = info.Name
		o.opts.version = info.Version
	}
}
