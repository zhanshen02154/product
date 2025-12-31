package wrapper

import (
	"context"
	"go-micro.dev/v4/broker"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

const deadLetterTopicKey = "DLQ"

// DeadLetterWrapper 死信队列
type DeadLetterWrapper struct {
	b broker.Broker
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
				return err
			}
			errStatus, ok := status.FromError(err)
			if !ok {
				logger.Errorf("failed to handler topic %v, error: %s; id: %s", msg.Topic(), err.Error(), msg.Header()["Micro-ID"])
				return err
			}
			switch errStatus.Code() {
			case codes.InvalidArgument:
				return err
			case codes.DataLoss:
				return err
			case codes.PermissionDenied:
				return err
			case codes.Unauthenticated:
				return err
			case codes.Aborted:
				return err
			case codes.NotFound:
				return err
			}

			dlMsg := broker.Message{
				Header: msg.Header(),
				Body:   msg.Body(),
			}
			dlMsg.Header["error"] = err.Error()
			topic := msg.Topic() + "DLQ"
			if err := w.b.Publish(topic, &dlMsg); err != nil {
				logger.Errorf("failed to publish to %s, error: %s", topic, err.Error())
			}
			return err
		}
	}
}

func NewDeadLetterWrapper(b broker.Broker) *DeadLetterWrapper {
	return &DeadLetterWrapper{
		b: b,
	}
}
