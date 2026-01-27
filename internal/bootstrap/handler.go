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
		ratelimit.NewHandlerWrapper(conf.Service.Qps),
		opentelemetry.NewHandlerWrapper(opentelemetry.WithTraceProvider(otel.GetTracerProvider())),
		prometheus.NewHandlerWrapper(prometheus.ServiceName(conf.Service.Name), prometheus.ServiceVersion(conf.Service.Version)),
		logWrapper.RequestLogWrapper,
	)
}
