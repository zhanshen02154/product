package main

import (
	"context"
	"fmt"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/util/log"
	service2 "github.com/zhanshen02154/product/internal/application/service"
	"github.com/zhanshen02154/product/internal/infrastructure"
	config2 "github.com/zhanshen02154/product/internal/infrastructure/config"
	"github.com/zhanshen02154/product/internal/infrastructure/persistence"
	gorm2 "github.com/zhanshen02154/product/internal/infrastructure/persistence/gorm"
	registry2 "github.com/zhanshen02154/product/internal/infrastructure/registry"
	"github.com/zhanshen02154/product/internal/intefaces/handler"
	"github.com/zhanshen02154/product/proto/product"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	confInfo, err := config2.LoadSystemConfig()
	if err != nil {
		panic(err)
	}
	// 注册到Consul
	consulRegistry := registry2.ConsulRegister(&confInfo.Consul)

	//链路追踪
	//tracer, io, err := common.NewTracer(cmd.SERVICE_NAME, cmd.TRACER_ADDR)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer io.Close()
	//opetracing2.SetGlobalTracer(tracer)

	db, err := persistence.InitDB(&confInfo.Database)
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	probeServer := infrastructure.NewProbeServer(":8080", db)
	probeServer.Start(ctx)

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
	)

	// Initialise service
	service.Init()

	productService := service2.NewProductApplicationService(txManager, productRepo)
	err = product.RegisterProductHandler(service.Server(), &handler.ProductHandler{ProductApplicationService: productService})
	if err != nil {
		log.Error(err)
	}

	// Run service
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-sigChan:
			cancel()
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err = service.Run(); err != nil {
			log.Fatal(err)
			cancel()
		}
	}()
	wg.Wait()
	probeServer.Wait()
}
