package bootstrap

import (
	"github.com/go-micro/plugins/v4/wrapper/monitoring/prometheus"
	"github.com/go-micro/plugins/v4/wrapper/trace/opentelemetry"
	"github.com/zhanshen02154/product/internal/config"
	"github.com/zhanshen02154/product/internal/infrastructure"
	"go-micro.dev/v4"
	"go.opentelemetry.io/otel"
)

func wrapSubscriber(conf *config.ServiceInfo, logWrapper *infrastructure.LogWrapper) micro.Option {
	return micro.WrapSubscriber(
		opentelemetry.NewSubscriberWrapper(opentelemetry.WithTraceProvider(otel.GetTracerProvider())),
		prometheus.NewSubscriberWrapper(prometheus.ServiceName(conf.Name), prometheus.ServiceVersion(conf.Version)),
		logWrapper.SubscribeWrapper(),
	)
}
