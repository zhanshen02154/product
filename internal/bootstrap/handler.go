package bootstrap

import (
	"github.com/go-micro/plugins/v4/wrapper/monitoring/prometheus"
	ratelimit "github.com/go-micro/plugins/v4/wrapper/ratelimiter/uber"
	"github.com/go-micro/plugins/v4/wrapper/trace/opentelemetry"
	"github.com/zhanshen02154/product/internal/config"
	"github.com/zhanshen02154/product/internal/infrastructure"
	"go-micro.dev/v4"
	"go.opentelemetry.io/otel"
)

func handlerWrapper(conf *config.ServiceInfo, logWrapper *infrastructure.LogWrapper) micro.Option {
	return micro.WrapHandler(
		ratelimit.NewHandlerWrapper(conf.Qps),
		opentelemetry.NewHandlerWrapper(opentelemetry.WithTraceProvider(otel.GetTracerProvider())),
		prometheus.NewHandlerWrapper(prometheus.ServiceName(conf.Name), prometheus.ServiceVersion(conf.Version)),
		logWrapper.RequestLogWrapper,
	)
}
