package wrapper

import (
	"context"
	"github.com/go-micro/plugins/v4/wrapper/trace/opentelemetry"
	"github.com/zhanshen02154/product/internal/infrastructure/event/monitor"
	"go-micro.dev/v4/broker"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"
	"go.opentelemetry.io/otel"
	tracecode "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"strings"
	"time"
)

const deadLetterTopicKey = "DLQ"

// DeadLetterWrapper 死信队列
type DeadLetterWrapper struct {
	b             broker.Broker
	traceProvicer trace.TracerProvider
}

// Wrapper 包装器操作
func (w *DeadLetterWrapper) Wrapper() server.SubscriberWrapper {
	return func(next server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, msg server.Message) error {
			err := next(ctx, msg)
			if err == nil {
				return nil
			}

			// 如果是死信队列则直接返回不再进入死信队列过程
			if strings.HasSuffix(msg.Topic(), deadLetterTopicKey) {
				return nil
			}
			errStatus, ok := status.FromError(err)
			if !ok {
				logger.Errorf("failed to handler topic %v, error: %s; id: %s", msg.Topic(), err.Error(), msg.Header()["Micro-ID"])
				return nil
			}

			// 根据errorCode判断 因日志已经记录所以直接返回nil
			switch errStatus.Code() {
			case codes.InvalidArgument:
				return nil
			case codes.DataLoss:
				return nil
			case codes.PermissionDenied:
				return nil
			case codes.Unauthenticated:
				return nil
			case codes.Aborted:
				return nil
			case codes.NotFound:
				return nil
			}
			spanOpts := []trace.SpanStartOption{
				trace.WithSpanKind(trace.SpanKindProducer),
			}
			topic := msg.Topic() + deadLetterTopicKey
			newCtx, span := opentelemetry.StartSpanFromContext(ctx, w.traceProvicer, "Pub to dead letter topic "+topic, spanOpts...)
			defer span.End()
			header := make(map[string]string)
			header["x-error"] = err.Error()
			for k, v := range msg.Header() {
				header[k] = v
			}
			header["Timestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
			dlMsg := broker.Message{
				Header: header,
				Body:   msg.Body(),
			}
			if err := w.b.Publish(topic, &dlMsg, broker.PublishContext(newCtx)); err != nil {
				logger.Errorf("failed to publish to %s, error: %s", topic, err.Error())
				span.SetStatus(tracecode.Error, err.Error())
				span.RecordError(err)
			} else {
				monitor.MessagesInFlight.WithLabelValues(topic, header["Source"], header["Schema_version"]).Inc()
			}

			// 一律返回nil让broker标记为成功
			return nil
		}
	}
}

func NewDeadLetterWrapper(b broker.Broker) *DeadLetterWrapper {
	return &DeadLetterWrapper{
		b:             b,
		traceProvicer: otel.GetTracerProvider(),
	}
}
