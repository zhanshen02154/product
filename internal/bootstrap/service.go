package bootstrap

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/go-micro/plugins/v4/broker/kafka"
	grpcclient "github.com/go-micro/plugins/v4/client/grpc"
	grpcserver "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/go-micro/plugins/v4/transport/grpc"
	"github.com/go-micro/plugins/v4/wrapper/monitoring/prometheus"
	ratelimit "github.com/go-micro/plugins/v4/wrapper/ratelimiter/uber"
	"github.com/go-micro/plugins/v4/wrapper/trace/opentelemetry"
	appservice "github.com/zhanshen02154/product/internal/application/service"
	"github.com/zhanshen02154/product/internal/config"
	"github.com/zhanshen02154/product/internal/infrastructure"
	"github.com/zhanshen02154/product/internal/infrastructure/event"
	"github.com/zhanshen02154/product/internal/infrastructure/event/monitor"
	"github.com/zhanshen02154/product/internal/infrastructure/event/wrapper"
	"github.com/zhanshen02154/product/internal/infrastructure/registry"
	"github.com/zhanshen02154/product/internal/intefaces/handler"
	"github.com/zhanshen02154/product/internal/intefaces/subscriber"
	"github.com/zhanshen02154/product/pkg/codec"
	"github.com/zhanshen02154/product/proto/product"
	"go-micro.dev/v4"
	broker2 "go-micro.dev/v4/broker"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"time"
)

func RunService(conf *config.SysConfig, serviceContext *infrastructure.ServiceContext, zapLogger *zap.Logger) error {
	// 注册到Consul
	consulRegistry := registry.ConsulRegister(conf.Consul)

	// 健康检查
	probeServer := infrastructure.NewProbeServer(conf.Service.HeathCheckAddr, serviceContext)

	monitorSvr := infrastructure.NewMonitorServer(":6060")

	// New Service
	logWrapper := infrastructure.NewLogWrapper(
		infrastructure.WithZapLogger(zapLogger),
		infrastructure.WithRequestSlowThreshold(conf.Service.RequestSlowThreshold),
		infrastructure.WithSubscribeSlowThreshold(conf.Broker.SubscribeSlowThreshold),
	)
	var eb event.Listener
	client := grpcclient.NewClient(
		grpcclient.PoolMaxIdle(100),
	)
	// 为 AsyncProducer 准备 channels，并把它们传给 kafka 插件
	// 使用与 Kafka 配置中相同的缓冲，减少短时写阻塞风险
	successChan := make(chan *sarama.ProducerMessage, conf.Broker.Kafka.ChannelBufferSize)
	errorChan := make(chan *sarama.ProducerError, conf.Broker.Kafka.ChannelBufferSize)
	broker := infrastructure.NewKafkaBroker(conf.Broker.Kafka, kafka.AsyncProducer(errorChan, successChan))
	broker.Init(broker2.ErrorHandler(wrapper.ErrorHandler(broker)))
	service := micro.NewService(
		micro.Server(grpcserver.NewServer(
			server.Name(conf.Service.Name),
			server.Version(conf.Service.Version),
			server.Address(conf.Service.Listen),
			server.Transport(grpc.NewTransport()),
			server.Registry(consulRegistry),
			server.RegisterTTL(time.Duration(conf.Consul.RegisterTtl)*time.Second),
			server.RegisterInterval(time.Duration(conf.Consul.RegisterInterval)*time.Second),
			grpcserver.Codec("application/grpc+dtm_raw", codec.NewDtmCodec()),
		)),
		micro.Client(client),
		//添加限流
		micro.WrapHandler(
			ratelimit.NewHandlerWrapper(conf.Service.Qps),
			opentelemetry.NewHandlerWrapper(opentelemetry.WithTraceProvider(otel.GetTracerProvider())),
			prometheus.NewHandlerWrapper(prometheus.ServiceName(conf.Service.Name), prometheus.ServiceVersion(conf.Service.Version)),
			logWrapper.RequestLogWrapper,
		),
		micro.AfterStart(func() error {
			if err := probeServer.Start(); err != nil {
				logger.Error("failed to start probe server: " + err.Error())
			}
			monitorSvr.Start()
			if eb != nil {
				eb.Start()
			}
			return nil
		}),
		micro.BeforeStop(func() error {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			logger.Info("Stopping monitor servers...")
			if err := monitorSvr.Close(shutdownCtx); err != nil {
				logger.Error("failed to close monitor servers" + err.Error())
			} else {
				logger.Info("Stopping monitor servers successfully")
			}
			if err := probeServer.Shutdown(shutdownCtx); err != nil {
				logger.Error("Failed to close probe server" + err.Error())
			} else {
				logger.Info("Successfully closed monitor servers")
			}
			serviceContext.Close()
			return nil
		}),
		micro.Broker(broker),
		micro.WrapClient(
			wrapper.NewMetaDataWrapper(conf.Service.Name, conf.Service.Version),
			opentelemetry.NewClientWrapper(opentelemetry.WithTraceProvider(otel.GetTracerProvider())),
			monitor.NewClientWrapper(monitor.WithName(conf.Service.Name), monitor.WithVersion(conf.Service.Version)),
		),
		micro.WrapSubscriber(
			logWrapper.SubscribeWrapper(),
			prometheus.NewSubscriberWrapper(prometheus.ServiceName(conf.Service.Name), prometheus.ServiceVersion(conf.Service.Version)),
			opentelemetry.NewSubscriberWrapper(opentelemetry.WithTraceProvider(otel.GetTracerProvider())),
		),
		micro.AfterStop(func() error {
			if eb != nil {
				eb.Close()
			}
			return nil
		}),
	)

	// 注册应用层服务及事件侦听器
	eb = event.NewListener(
		event.WithProducerChannels(successChan, errorChan),
		event.WithServiceName(conf.Service.Name),
		event.WithServiceVersion(conf.Service.Version),
		event.WrapPublishCallback(
			event.NewDeadletterWrapper(event.WithBroker(broker), event.WithTracer(otel.GetTracerProvider()), event.WithServiceInfo(conf.Service)),
			event.NewPublicCallbackLogWrapper(
				event.WithLogger(zapLogger),
				event.WithTimeThreshold(conf.Broker.PublishTimeThreshold),
			),
			event.NewTracerWrapper(event.WithTracerProvider(otel.GetTracerProvider())),
		),
	)
	event.RegisterPublisher(conf.Broker, eb, service.Client())
	productService := appservice.NewProductApplicationService(serviceContext, eb)

	paymentEventHandler := subscriber.NewPaymentEventHandler(productService)
	paymentEventHandler.RegisterSubscriber(service.Server())

	err := product.RegisterProductHandler(service.Server(), handler.NewProductHandler(productService))
	if err != nil {
		return err
	}

	// Run service
	if err = service.Run(); err != nil {
		return err
	}

	return nil
}
