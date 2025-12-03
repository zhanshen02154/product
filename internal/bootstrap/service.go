package bootstrap

import (
	"context"
	grpcserver "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/go-micro/plugins/v4/transport/grpc"
	ratelimit "github.com/go-micro/plugins/v4/wrapper/ratelimiter/uber"
	appservice "github.com/zhanshen02154/product/internal/application/service"
	"github.com/zhanshen02154/product/internal/config"
	"github.com/zhanshen02154/product/internal/infrastructure"
	"github.com/zhanshen02154/product/internal/infrastructure/broker/kafka"
	event2 "github.com/zhanshen02154/product/internal/infrastructure/event"
	"github.com/zhanshen02154/product/internal/infrastructure/event/wrapper"
	"github.com/zhanshen02154/product/internal/infrastructure/registry"
	innerserver "github.com/zhanshen02154/product/internal/infrastructure/server"
	event3 "github.com/zhanshen02154/product/internal/intefaces/event"
	"github.com/zhanshen02154/product/internal/intefaces/handler"
	"github.com/zhanshen02154/product/pkg/codec"
	"github.com/zhanshen02154/product/proto/product"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"
	"time"
)

func RunService(conf *config.SysConfig, serviceContext *infrastructure.ServiceContext) error {
	// 注册到Consul
	consulRegistry := registry.ConsulRegister(conf.Consul)

	//链路追踪
	//tracer, io, err := common.NewTracer(cmd.SERVICE_NAME, cmd.TRACER_ADDR)
	//if err != nil {
	//	logger.Error(err)
	//}
	//defer io.Close()
	//opetracing2.SetGlobalTracer(tracer)

	// 健康检查
	probeServer := innerserver.NewProbeServer(conf.Service.HeathCheckAddr, serviceContext)
	if err := probeServer.Start(); err != nil {
		logger.Error("健康检查服务器启动失败")
	}

	var pprofSrv *innerserver.PprofServer
	if conf.Service.Debug {
		pprofSrv = innerserver.NewPprofServer(":6060")
	}

	// New Service
	broker := kafka.NewKafkaBroker(conf.Broker.Kafka)
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
		//micro.WrapHandler(opentracing.NewHandlerWrapper(opetracing2.GlobalTracer())),
		//添加限流
		micro.WrapHandler(ratelimit.NewHandlerWrapper(conf.Service.Qps)),
		//micro.WrapHandler(opentracing.NewHandlerWrapper(opetracing2.GlobalTracer())),
		micro.AfterStart(func() error {
			pprofSrv.Start()
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
		micro.WrapClient(wrapper.NewMetaDataWrapper(conf.Service.Name, conf.Service.Version)),
	)

	// 注册应用层服务及事件侦听器
	eb := event2.NewListener(service.Client())
	productService := appservice.NewProductApplicationService(serviceContext, eb)
	registerPublisher(conf.Broker, eb)
	defer eb.Close()

	productEventHandler := event3.NewHandler(productService)
	// 注册订阅事件
	if len(conf.Broker.Subscriber) > 0 {
		for i := range conf.Broker.Subscriber {
			if err := micro.RegisterSubscriber(conf.Broker.Subscriber[i], service.Server(), productEventHandler); err != nil {
				logger.Error("failed to register subsriber: ", err)
				continue
			}
		}
	}

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

// 注册发布事件
func registerPublisher(conf *config.Broker, eb event2.Bus) {
	if len(conf.Publisher) > 0 {
		for i := range conf.Publisher {
			eb.Register(conf.Publisher[i])
		}
	}
}
