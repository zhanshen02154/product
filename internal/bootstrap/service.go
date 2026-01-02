package bootstrap

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/go-micro/plugins/v4/broker/kafka"
	grpcclient "github.com/go-micro/plugins/v4/client/grpc"
	grpcserver "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/go-micro/plugins/v4/transport/grpc"
	ratelimit "github.com/go-micro/plugins/v4/wrapper/ratelimiter/uber"
	"github.com/go-micro/plugins/v4/wrapper/trace/opentelemetry"
	appservice "github.com/zhanshen02154/product/internal/application/service"
	"github.com/zhanshen02154/product/internal/config"
	"github.com/zhanshen02154/product/internal/infrastructure"
	"github.com/zhanshen02154/product/internal/infrastructure/event"
	"github.com/zhanshen02154/product/internal/infrastructure/event/wrapper"
	"github.com/zhanshen02154/product/internal/infrastructure/registry"
	"github.com/zhanshen02154/product/internal/intefaces/handler"
	"github.com/zhanshen02154/product/internal/intefaces/subscriber"
	"github.com/zhanshen02154/product/pkg/codec"
	"github.com/zhanshen02154/product/proto/product"
	"go-micro.dev/v4"
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
	if err := probeServer.Start(); err != nil {
		logger.Error("健康检查服务器启动失败")
	}

	var pprofSrv *infrastructure.PprofServer
	if conf.Service.Debug {
		pprofSrv = infrastructure.NewPprofServer(":6060")
	}

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
	deadLetterWrapper := wrapper.NewDeadLetterWrapper(broker)
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
			logWrapper.RequestLogWrapper,
		),
		micro.AfterStart(func() error {
			pprofSrv.Start()
			if eb != nil {
				eb.Start()
			}
			return nil
		}),
		micro.BeforeStop(func() error {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			logger.Info("收到关闭信号，正在停止健康检查服务器...")
			err := probeServer.Shutdown(shutdownCtx)
			if err != nil {
				return err
			}
			if conf.Service.Debug {
				if err := pprofSrv.Close(shutdownCtx); err != nil {
					logger.Error("pprof服务器关闭错误: ", err)
				}
			}
			return nil
		}),
		micro.Broker(broker),
		micro.WrapClient(
			wrapper.NewMetaDataWrapper(conf.Service.Name, conf.Service.Version),
			opentelemetry.NewClientWrapper(opentelemetry.WithTraceProvider(otel.GetTracerProvider())),
		),
		micro.WrapSubscriber(
			opentelemetry.NewSubscriberWrapper(opentelemetry.WithTraceProvider(otel.GetTracerProvider())),
			deadLetterWrapper.Wrapper(),
			logWrapper.SubscribeWrapper(),
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
		event.WithBroker(broker),
		event.WithClient(service.Client()),
		event.WithLogger(zapLogger),
		event.WithPulishTimeThreshold(conf.Broker.Kafka.Producer.PublishTimeThreshold),
		event.WithProducerChannels(successChan, errorChan),
	)
	event.RegisterPublisher(conf.Broker, eb)
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
