package main

import (
	"fmt"
	"git.imooc.com/zhanshen1614/product/handler"
	"git.imooc.com/zhanshen1614/product/internal/config"
	service2 "git.imooc.com/zhanshen1614/product/internal/domain/service"
	config2 "git.imooc.com/zhanshen1614/product/internal/infrastructure/config"
	gorm2 "git.imooc.com/zhanshen1614/product/internal/infrastructure/persistence/gorm"
	registry2 "git.imooc.com/zhanshen1614/product/internal/infrastructure/registry"
	"git.imooc.com/zhanshen1614/product/proto/product"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/util/log"
	"net/http"
	"net/url"
	"time"
	_ "time/tzdata"
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

	db, err := initDB(confInfo)
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}
	defer func() {
		if db != nil { // 关键检查
			db.Close()
		}
	}()
	rp := gorm2.NewProductRepository(db)

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

	go func() {
		// livenessProbe
		http.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("OK"))
		})

		// readinessProbe
		http.HandleFunc("/ready", func(writer http.ResponseWriter, request *http.Request) {
			if err := db.DB().Ping(); err != nil {
				writer.WriteHeader(http.StatusServiceUnavailable)
				writer.Write([]byte("Not Ready"))
			} else {
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte("Ok"))
			}
		})
		err = http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatalf("check status http api error: %v", err)
		} else {
			log.Info("listen http server on: 8080")
		}
	}()

	productService := service2.NewProductDataService(rp)
	err = product.RegisterProductHandler(service.Server(), &handler.Product{ProductDataService: productService})
	if err != nil {
		log.Error(err)
	}

	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}

// 加载数据库
func initDB(confInfo *config.SysConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s",
		confInfo.Database.User,
		confInfo.Database.Password,
		confInfo.Database.Host,
		confInfo.Database.Port,
		confInfo.Database.Database,
		confInfo.Database.Charset,
		url.QueryEscape(confInfo.Database.Loc),
	)
	fmt.Println(dsn)
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	sqlDB := db.DB()
	if sqlDB == nil {
		return nil, fmt.Errorf("获取SQL DB失败: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(1000)
	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 验证连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接验证失败: %w", err)
	}

	log.Info("数据库连接成功")
	return db, nil
}
