package event

import (
	"context"
	"github.com/go-micro/plugins/v4/wrapper/trace/opentelemetry"
	"go-micro.dev/v4/broker"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type TraceOptions struct {
	traceProvider trace.TracerProvider
}

type TraceOption func(opts *TraceOptions)

// WithTracerProvider 使用TracerProvider
func WithTracerProvider(provider trace.TracerProvider) TraceOption {
	return func(opts *TraceOptions) {
		opts.traceProvider = provider
	}
}

// NewTracerWrapper 创建链路追踪包装器
func NewTracerWrapper(traceOpts ...TraceOption) PublishCallbackWrapper {
	opts := TraceOptions{}
	for _, opt := range traceOpts {
		opt(&opts)
	}
	return func(next PublishCallbackFunc) PublishCallbackFunc {
		return func(ctx context.Context, msg *broker.Message, err error) {
			spanOpts := []trace.SpanStartOption{
				trace.WithSpanKind(trace.SpanKindProducer),
			}
			newCtx, span := opentelemetry.StartSpanFromContext(ctx, opts.traceProvider, "Kafka publish callback", spanOpts...)
			next(newCtx, msg, err)
			defer span.End()
			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
			} else {
				partition := getPartition(ctx)
				offset := getOffset(ctx)
				if partition >= 0 && offset > 0 {
					span.SetAttributes(
						attribute.Int64("kafka.offset", offset),
						attribute.Int("kafka.partition", int(partition)),
						attribute.String("kafka.topic", msg.Header["Micro-Topic"]),
					)
				}
			}
			return
		}
	}
}
