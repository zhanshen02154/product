package bootstrap

import (
	"github.com/go-micro/plugins/v4/wrapper/monitoring/prometheus"
	ratelimit "github.com/go-micro/plugins/v4/wrapper/ratelimiter/uber"
	"github.com/go-micro/plugins/v4/wrapper/trace/opentelemetry"
	"github.com/zhanshen02154/product/internal/config"
	"github.com/zhanshen02154/product/internal/infrastructure"
	"go-micro.dev/v4"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func handlerWrapper(conf *config.SysConfig, l *zap.Logger) micro.Option {
	logWrapper := infrastructure.NewLogWrapper(
		infrastructure.WithZapLogger(l),
		infrastructure.WithRequestSlowThreshold(conf.Service.RequestSlowThreshold),
		infrastructure.WithSubscribeSlowThreshold(conf.Broker.SubscribeSlowThreshold),
	)

	return micro.WrapHandler(
		logWrapper.RequestLogWrapper,
		prometheus.NewHandlerWrapper(prometheus.ServiceName(conf.Service.Name), prometheus.ServiceVersion(conf.Service.Version)),
		opentelemetry.NewHandlerWrapper(opentelemetry.WithTraceProvider(otel.GetTracerProvider())),
		ratelimit.NewHandlerWrapper(conf.Service.Qps),
	)
}
