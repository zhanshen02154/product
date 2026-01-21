package bootstrap

import (
	"github.com/go-micro/plugins/v4/wrapper/monitoring/prometheus"
	"github.com/go-micro/plugins/v4/wrapper/trace/opentelemetry"
	"github.com/zhanshen02154/product/internal/config"
	"github.com/zhanshen02154/product/internal/infrastructure"
	"go-micro.dev/v4"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func wrapSubscriber(logger *zap.Logger, conf *config.SysConfig) micro.Option {
	logWrapper := infrastructure.NewLogWrapper(
		infrastructure.WithZapLogger(logger),
		infrastructure.WithRequestSlowThreshold(conf.Service.RequestSlowThreshold),
		infrastructure.WithSubscribeSlowThreshold(conf.Broker.SubscribeSlowThreshold),
	)
	return micro.WrapSubscriber(
		opentelemetry.NewSubscriberWrapper(opentelemetry.WithTraceProvider(otel.GetTracerProvider())),
		prometheus.NewSubscriberWrapper(prometheus.ServiceName(conf.Service.Name), prometheus.ServiceVersion(conf.Service.Version)),
		logWrapper.SubscribeWrapper(),
	)
}
