package event

import (
	"context"
	metadata2 "github.com/zhanshen02154/product/pkg/metadata"
	"go-micro.dev/v4/broker"
	"go.uber.org/zap"
	"strconv"
	"time"
)

// 事件侦听器配置
type listenerLoggerOptions struct {
	publishTimeThreshold time.Duration
	logger               *zap.Logger
}

type LogOption func(wrapper *listenerLoggerOptions)

func NewPublicCallbackLogWrapper(opts ...LogOption) PublishCallbackWrapper {
	logOpts := listenerLoggerOptions{}
	for _, o := range opts {
		o(&logOpts)
	}
	return func(next PublishCallbackFunc) PublishCallbackFunc {
		return func(ctx context.Context, msg *broker.Message, err error) {
			next(ctx, msg, err)
			var duration time.Duration
			currentTime := time.Now()
			if _, ok := msg.Header["Timestamp"]; ok {
				startTimeInt, err := strconv.ParseInt(msg.Header["Timestamp"], 10, 64)
				if err != nil {
					return
				}
				startTime := time.UnixMilli(startTimeInt)
				duration = currentTime.Sub(startTime)
			} else {
				duration = 0
			}

			logFields := make([]zap.Field, 0, 15)
			logFields = append(logFields,
				zap.String("type", "publish"),
				zap.String("trace_id", metadata2.GetTraceIdFromSpan(ctx)),
				zap.String("topic", msg.Header["Micro-Topic"]),
				zap.String("event_id", msg.Header["Event_id"]),
				zap.String("source", msg.Header["Source"]),
				zap.String("schema_version", msg.Header["Schema_version"]),
				zap.Int64("published_at", currentTime.Unix()),
				zap.String("remote", msg.Header["Remote"]),
				zap.String("accept_encoding", msg.Header["Accept-Encoding"]),
				zap.Int64("duration", duration.Milliseconds()),
			)
			if pKey, ok := msg.Header[partitionKey]; ok {
				logFields = append(logFields, zap.String("key", pKey))
			} else {
				logFields = append(logFields, zap.String("key", ""))
			}
			switch {
			case err != nil:
				logOpts.logger.Error("Publish event failed: "+err.Error(), logFields...)
			case err == nil && duration > logOpts.publishTimeThreshold && duration > 0:
				logFields = append(logFields, zap.String("stacktrace", ""))
				logOpts.logger.Warn("Publish event slow", logFields...)
			default:
				logFields = append(logFields, zap.String("stacktrace", ""))
				logOpts.logger.Info("Publish event success", logFields...)
			}
		}
	}
}

func WithLogger(l *zap.Logger) LogOption {
	return func(opts *listenerLoggerOptions) {
		opts.logger = l
	}
}

func WithTimeThreshold(times int64) LogOption {
	return func(opts *listenerLoggerOptions) {
		opts.publishTimeThreshold = time.Duration(times) * time.Millisecond
	}
}
