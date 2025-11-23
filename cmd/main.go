package main

import (
	"context"
	"fmt"
	"github.com/micro/go-micro/v2"
	config2 "github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/config/encoder/yaml"
	"github.com/micro/go-micro/v2/config/source"
	"github.com/micro/go-micro/v2/util/log"
	"github.com/micro/go-plugins/config/source/consul/v2"
	service2 "github.com/zhanshen02154/product/internal/application/service"
	configstruct "github.com/zhanshen02154/product/internal/config"
	"github.com/zhanshen02154/product/internal/infrastructure"
	"github.com/zhanshen02154/product/internal/infrastructure/persistence"
	gorm2 "github.com/zhanshen02154/product/internal/infrastructure/persistence/gorm"
	registry2 "github.com/zhanshen02154/product/internal/infrastructure/registry"
	"github.com/zhanshen02154/product/internal/intefaces/handler"
	"github.com/zhanshen02154/product/pkg/env"
	"github.com/zhanshen02154/product/proto/product"
	"time"
)

func main() {
	// 从consul获取配置
	consulHost := env.GetEnv("CONSUL_HOST", "192.168.83.131")
	consulPort := env.GetEnv("CONSUL_PORT", "8500")
	consulPrefix := env.GetEnv("CONSUL_PREFIX", "product")
	consulSource := consul.NewSource(
		// Set configuration address
		consul.WithAddress(fmt.Sprintf("%s:%s", consulHost, consulPort)),
		//前缀 默认：product
		consul.WithPrefix(consulPrefix),
		//consul.StripPrefix(true),
		source.WithEncoder(yaml.NewEncoder()),
	)
	configInfo, err := config2.NewConfig()
	defer func() {
		err = configInfo.Close()
		if err != nil {
			log.Error(err)
			return
		}
	}()
	if err != nil {
		log.Error(err)
		return
	}
	err = configInfo.Load(consulSource)
	if err != nil {
		log.Error(err)
		return
	}

	var confInfo configstruct.SysConfig
	if err = configInfo.Get(consulPrefix).Scan(&confInfo); err != nil {
		log.Error(err)
		return
	}

	// 注册到Consul
	consulRegistry := registry2.ConsulRegister(&confInfo.Consul)

	//链路追踪
	//tracer, io, err := common.NewTracer(cmd.SERVICE_NAME, cmd.TRACER_ADDR)
	//if err != nil {
	//	log.Error(err)
	//}
	//defer io.Close()
	//opetracing2.SetGlobalTracer(tracer)

	db, err := persistence.InitDB(&confInfo.Database)
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	// 健康检查
	probeServer := infrastructure.NewProbeServer(confInfo.Service.HeathCheckAddr, db)
	if err = probeServer.Start(); err != nil {
		log.Errorf("健康检查服务器启动失败")
	}

	txManager := gorm2.NewGormTransactionManager(db)
	productRepo := gorm2.NewProductRepository(db)

	// New Service
	service := micro.NewService(
		micro.Name(confInfo.Service.Name),
		micro.Version(confInfo.Service.Version),
		micro.Address(fmt.Sprintf(":%d", confInfo.Service.Port)),
		micro.Registry(consulRegistry),
		micro.RegisterTTL(time.Duration(confInfo.Consul.RegisterTtl)*time.Second),
		micro.RegisterInterval(time.Duration(confInfo.Consul.RegisterInterval)*time.Second),
		//micro.WrapHandler(opentracing.NewHandlerWrapper(opetracing2.GlobalTracer())),
		micro.BeforeStop(func() error {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			log.Info("收到关闭信号，正在停止健康检查服务器...")
			err = probeServer.Shutdown(shutdownCtx)
			if err != nil {
				return err
			}
			sqlDB, err := db.DB()
			if err != nil {
				log.Errorf("failed to close database: %v", err)
				return err
			}
			if err = sqlDB.Ping(); err == nil {
				err1 := sqlDB.Close()
				if err1 != nil {
					log.Infof("关闭GORM连接失败： %v", err1)
					return err1
				} else {
					log.Info("GORM数据库连接已关闭")
				}
			}
			return nil
		}),
	)

	// Initialise service
	//service.Init()

	productService := service2.NewProductApplicationService(txManager, productRepo)
	err = product.RegisterProductHandler(service.Server(), &handler.ProductHandler{ProductApplicationService: productService})
	if err != nil {
		log.Error(err)
		return
	}

	// Run service
	if err = service.Run(); err != nil {
		log.Error(err)
	}
}
