package main

import (
	"fmt"
	"github.com/go-micro/plugins/v4/config/source/consul"
	"github.com/zhanshen02154/product/internal/bootstrap"
	configstruct "github.com/zhanshen02154/product/internal/config"
	"github.com/zhanshen02154/product/internal/infrastructure"
	"github.com/zhanshen02154/product/pkg/env"
	config2 "go-micro.dev/v4/config"
	"go-micro.dev/v4/logger"
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

	serviceContext, err := infrastructure.NewServiceContext(&confInfo)
	defer serviceContext.Close()
	if err != nil {
		logger.Error("error to load service context: ", err)
	}

	if err := bootstrap.RunService(&confInfo, serviceContext); err != nil {
		logger.Error("failed to start service: ", err)
	}
}
