package wrapper

import (
	"context"
	"github.com/zhanshen02154/product/internal/infrastructure/event/monitor"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/metadata"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"strconv"
	"time"

	"github.com/go-micro/plugins/v4/wrapper/trace/opentelemetry"
	"go-micro.dev/v4/broker"
	"go.opentelemetry.io/otel/trace"
)

const deadLetterTopicKey = "DLQ"

// deadLetterHandler 死信队列
type deadLetterHandler struct {
	b             broker.Broker
	traceProvicer trace.TracerProvider
}

// ErrorHandler 错误处理
func ErrorHandler(b broker.Broker) broker.Handler {
	options := &deadLetterHandler{
		b:             b,
		traceProvicer: otel.GetTracerProvider(),
	}

	return func(event broker.Event) error {
		topic := event.Topic() + deadLetterTopicKey
		if v, ok := event.Message().Header["Traceparent"]; ok {
			event.Message().Header["traceparent"] = v
		}
		ctx := context.TODO()
		ctx = metadata.NewContext(ctx, event.Message().Header)
		err := options.publishDeadLetter(ctx, topic, event.Message(), event.Error())
		if err != nil {
			logger.Errorf("failed to publish to %s, error: %s", topic, err.Error())
		} else {
			monitor.MessagesInFlight.WithLabelValues(topic, event.Message().Header["Source"], event.Message().Header["Schema_version"]).Inc()
		}
		if err := event.Ack(); err != nil {
			logger.Errorf("failed to ack to %s, error: %s", topic, err.Error())
		}
		return err
	}
}

func (w *deadLetterHandler) publishDeadLetter(ctx context.Context, topic string, msg *broker.Message, err error) error {
	header := make(map[string]string)
	header["x-error"] = err.Error()
	for k, v := range msg.Header {
		header[k] = v
	}
	header["Timestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	dlMsg := broker.Message{
		Header: header,
		Body:   msg.Body,
	}
	spanOpts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindProducer),
	}
	newCtx, span := opentelemetry.StartSpanFromContext(ctx, w.traceProvicer, "Pub to deadletter topic "+topic, spanOpts...)
	defer span.End()
	pErr := w.b.Publish(topic, &dlMsg, broker.PublishContext(newCtx))
	if pErr != nil {
		span.SetStatus(codes.Error, pErr.Error())
		span.RecordError(pErr)
		return pErr
	}
	return nil
}
