package main

import (
	"context"
	"fmt"
	"github.com/go-micro/plugins/v4/config/source/consul"
	grpc2 "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/go-micro/plugins/v4/transport/grpc"
	ratelimit "github.com/go-micro/plugins/v4/wrapper/ratelimiter/uber"
	service2 "github.com/zhanshen02154/product/internal/application/service"
	configstruct "github.com/zhanshen02154/product/internal/config"
	"github.com/zhanshen02154/product/internal/infrastructure"
	registry2 "github.com/zhanshen02154/product/internal/infrastructure/registry"
	server2 "github.com/zhanshen02154/product/internal/infrastructure/server"
	"github.com/zhanshen02154/product/internal/intefaces/handler"
	"github.com/zhanshen02154/product/pkg/codec"
	"github.com/zhanshen02154/product/pkg/env"
	"github.com/zhanshen02154/product/proto/product"
	"go-micro.dev/v4"
	config2 "go-micro.dev/v4/config"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"
	"time"
)

func main() {
	// 从consul获取配置
	consulHost := env.GetEnv("CONSUL_HOST", "192.168.83.131")
	consulPort := env.GetEnv("CONSUL_PORT", "8500")
	consulPrefix := env.GetEnv("CONSUL_PREFIX", "/micro/")
	consulSource := consul.NewSource(
		// Set configuration address
		consul.WithAddress(fmt.Sprintf("%s:%s", consulHost, consulPort)),
		//前缀 默认：product
		consul.WithPrefix(consulPrefix),
		consul.StripPrefix(true),
	)
	configInfo, err := config2.NewConfig()
	defer func() {
		err = configInfo.Close()
		if err != nil {
			logger.Error(err)
			return
		}
	}()
	if err != nil {
		logger.Error(err)
		return
	}
	err = configInfo.Load(consulSource)
	if err != nil {
		logger.Error(err)
		return
	}

	var confInfo configstruct.SysConfig
	if err = configInfo.Get("product").Scan(&confInfo); err != nil {
		logger.Error(err)
		return
	}

	// 注册到Consul
	consulRegistry := registry2.ConsulRegister(confInfo.Consul)

	//链路追踪
	//tracer, io, err := common.NewTracer(cmd.SERVICE_NAME, cmd.TRACER_ADDR)
	//if err != nil {
	//	logger.Error(err)
	//}
	//defer io.Close()
	//opetracing2.SetGlobalTracer(tracer)

	serviceContext, err := infrastructure.NewServiceContext(&confInfo)
	defer serviceContext.Close()
	if err != nil {
		logger.Errorf("error to load service context: %s", err)
		return
	}

	// 健康检查
	probeServer := server2.NewProbeServer(confInfo.Service.HeathCheckAddr, serviceContext)
	if err := probeServer.Start(); err != nil {
		logger.Error("健康检查服务器启动失败")
	}

	var pprofSrv *server2.PprofServer
	if confInfo.Service.Debug {
		pprofSrv = server2.NewPprofServer(":6060")
	}

	// New Service
	service := micro.NewService(
		micro.Server(grpc2.NewServer(
			server.Name(confInfo.Service.Name),
			server.Version(confInfo.Service.Version),
			server.Address(confInfo.Service.Listen),
			server.Transport(grpc.NewTransport()),
			server.Registry(consulRegistry),
			server.RegisterTTL(time.Duration(confInfo.Consul.RegisterTtl)*time.Second),
			server.RegisterInterval(time.Duration(confInfo.Consul.RegisterInterval)*time.Second),
			grpc2.Codec("application/grpc+dtm_raw", codec.NewDtmCodec()),
		)),
		//micro.WrapHandler(opentracing.NewHandlerWrapper(opetracing2.GlobalTracer())),
		//添加限流
		micro.WrapHandler(ratelimit.NewHandlerWrapper(confInfo.Service.Qps)),
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
			if confInfo.Service.Debug {
				if err := pprofSrv.Close(shutdownCtx); err != nil {
					logger.Error("pprof服务器关闭错误: ", err)
				}
			}
			return nil
		}),
	)

	// Initialise service
	//service.Init()

	productService := service2.NewProductApplicationService(serviceContext)
	err = product.RegisterProductHandler(service.Server(), handler.NewProductHandler(productService))
	if err != nil {
		logger.Error(err)
		return
	}

	// Run service
	if err = service.Run(); err != nil {
		logger.Error(err)
	}
}
