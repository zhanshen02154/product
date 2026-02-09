package bootstrap

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/go-micro/plugins/v4/broker/kafka"
	grpcclient "github.com/go-micro/plugins/v4/client/grpc"
	"github.com/go-micro/plugins/v4/wrapper/trace/opentelemetry"
	appservice "github.com/zhanshen02154/product/internal/application/service"
	"github.com/zhanshen02154/product/internal/config"
	"github.com/zhanshen02154/product/internal/infrastructure"
	"github.com/zhanshen02154/product/internal/infrastructure/event"
	"github.com/zhanshen02154/product/internal/infrastructure/event/monitor"
	"github.com/zhanshen02154/product/internal/infrastructure/event/wrapper"
	"github.com/zhanshen02154/product/internal/intefaces/handler"
	"github.com/zhanshen02154/product/internal/intefaces/subscriber"
	"github.com/zhanshen02154/product/proto/product"
	"go-micro.dev/v4"
	broker2 "go-micro.dev/v4/broker"
	"go-micro.dev/v4/logger"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"time"
)

func RunService(conf *config.SysConfig, serviceContext *infrastructure.ServiceContext, zapLogger *zap.Logger) error {
	// 健康检查
	probeServer := infrastructure.NewProbeServer(conf.Service.HeathCheckAddr, serviceContext)

	monitorSvr := infrastructure.NewMonitorServer(":6060")

	// New Service
	var eb event.Listener
	client := grpcclient.NewClient(
		grpcclient.PoolMaxIdle(100),
	)
	// 为 AsyncProducer 准备 channels，并把它们传给 kafka 插件
	// 使用与 Kafka 配置中相同的缓冲，减少短时写阻塞风险
	successChan := make(chan *sarama.ProducerMessage, conf.Broker.Kafka.ChannelBufferSize)
	errorChan := make(chan *sarama.ProducerError, conf.Broker.Kafka.ChannelBufferSize)
	broker := infrastructure.NewKafkaBroker(conf.Broker.Kafka, kafka.AsyncProducer(errorChan, successChan))
	broker2.DefaultBroker = broker
	logWrapper := infrastructure.NewLogWrapper(
		infrastructure.WithZapLogger(zapLogger),
		infrastructure.WithRequestSlowThreshold(conf.Service.RequestSlowThreshold),
		infrastructure.WithSubscribeSlowThreshold(conf.Broker.SubscribeSlowThreshold),
	)
	service := micro.NewService(
		newServer(conf),
		micro.Client(client),
		handlerWrapper(conf.Service, logWrapper),
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
		wrapSubscriber(conf, logWrapper),
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
			event.NewTracerWrapper(event.WithTracerProvider(otel.GetTracerProvider())),
			event.NewDeadletterWrapper(event.WithTracer(otel.GetTracerProvider()), event.WithServiceInfo(conf.Service)),
			event.NewPublicCallbackLogWrapper(
				event.WithLogger(zapLogger),
				event.WithTimeThreshold(conf.Broker.PublishTimeThreshold),
			),
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
