package retry

import (
	"context"
	"github.com/cenkalti/backoff/v4"
	metadatahelper "github.com/zhanshen02154/product/pkg/metadata"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/metadata"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type Policy interface {
	Execute(ctx context.Context, fn func() error) error
}

type exponentialBackOff struct {
	opts *options
}

func (r *exponentialBackOff) Execute(ctx context.Context, fn func() error) error {
	err := fn()
	if err == nil {
		return nil
	}
	if !r.checkRetry(err) {
		return err
	}
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = r.opts.initialInterval
	expBackoff.MaxInterval = r.opts.maxInterval
	expBackoff.MaxElapsedTime = r.opts.maxElapsedTime
	backoffRetry := backoff.WithMaxRetries(expBackoff, r.opts.maxRetries)
	if err := backoff.RetryNotify(fn, backoffRetry, r.notify(ctx)); err != nil {
		return err
	}
	return nil
}

func (r *exponentialBackOff) notify(ctx context.Context) backoff.Notify {
	return func(err error, duration time.Duration) {
		topic, ok := metadata.Get(ctx, "Micro-Topic")
		if !ok {
			logger.Error("subscriber topic does not exist")
			return
		}

		switch {
		case err != nil:
			r.opts.logger.Error(topic+" subsriber handler retry failed: "+err.Error(),
				zap.String("type", "subscribe"),
				zap.String("trace_id", metadatahelper.GetTraceIdFromSpan(ctx)),
				zap.String("event_id", metadatahelper.GetValueFromMetadata(ctx, "Event_id")),
				zap.String("topic", topic),
				zap.String("source", metadatahelper.GetValueFromMetadata(ctx, "Source")),
				zap.String("schema_version", metadatahelper.GetValueFromMetadata(ctx, "Schema_version")),
				zap.String("grpc_accept_encoding", metadatahelper.GetValueFromMetadata(ctx, "Grpc-Accept-Encoding")),
				zap.String("remote", metadatahelper.GetValueFromMetadata(ctx, "Remote")),
				zap.String("accept_encoding", metadatahelper.GetValueFromMetadata(ctx, "Accept-Encoding")),
				zap.String("key", metadatahelper.GetValueFromMetadata(ctx, "Pkey")),
				zap.Int64("duration", duration.Milliseconds()),
			)
		default:
			r.opts.logger.Info(topic+" subsriber handler retry success",
				zap.String("type", "subscribe"),
				zap.String("trace_id", metadatahelper.GetTraceIdFromSpan(ctx)),
				zap.String("event_id", metadatahelper.GetValueFromMetadata(ctx, "Event_id")),
				zap.String("topic", topic),
				zap.String("source", metadatahelper.GetValueFromMetadata(ctx, "Source")),
				zap.String("schema_version", metadatahelper.GetValueFromMetadata(ctx, "Schema_version")),
				zap.String("grpc_accept_encoding", metadatahelper.GetValueFromMetadata(ctx, "Grpc-Accept-Encoding")),
				zap.String("remote", metadatahelper.GetValueFromMetadata(ctx, "Remote")),
				zap.String("accept_encoding", metadatahelper.GetValueFromMetadata(ctx, "Accept-Encoding")),
				zap.String("key", metadatahelper.GetValueFromMetadata(ctx, "Pkey")),
				zap.Int64("duration", duration.Milliseconds()),
			)
		}
	}
}

// 检查是否需要重试
func (r *exponentialBackOff) checkRetry(err error) bool {
	if err == nil {
		return false
	}
	errStatus, ok := status.FromError(err)
	if !ok {
		return false
	}

	switch errStatus.Code() {
	case codes.InvalidArgument:
		return false
	case codes.DataLoss:
		return false
	case codes.PermissionDenied:
		return false
	case codes.Unauthenticated:
		return false
	case codes.FailedPrecondition:
		return false
	case codes.NotFound:
		return false
	case codes.Unimplemented:
		return false
	case codes.Internal:
		return false
	case codes.AlreadyExists:
		return false
	case codes.ResourceExhausted:
		return false
	}

	return true
}

func NewRetryPolicy(opts ...Option) Policy {
	newOptions := options{}
	for _, opt := range opts {
		opt(&newOptions)
	}

	return &exponentialBackOff{
		opts: &newOptions,
	}
}
