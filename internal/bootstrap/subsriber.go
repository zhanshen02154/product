package bootstrap

import (
	"github.com/go-micro/plugins/v4/wrapper/monitoring/prometheus"
	"github.com/go-micro/plugins/v4/wrapper/trace/opentelemetry"
	"github.com/zhanshen02154/product/internal/config"
	"github.com/zhanshen02154/product/internal/infrastructure"
	"go-micro.dev/v4"
	"go.opentelemetry.io/otel"
)

func wrapSubscriber(conf *config.SysConfig, logWrapper *infrastructure.LogWrapper) micro.Option {
	return micro.WrapSubscriber(
		opentelemetry.NewSubscriberWrapper(opentelemetry.WithTraceProvider(otel.GetTracerProvider())),
		prometheus.NewSubscriberWrapper(prometheus.ServiceName(conf.Service.Name), prometheus.ServiceVersion(conf.Service.Version)),
		logWrapper.SubscribeWrapper(),
	)
}
