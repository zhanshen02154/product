package main

import (
	"github.com/zhanshen02154/product/internal/bootstrap"
	configstruct "github.com/zhanshen02154/product/internal/config"
	"go-micro.dev/v4/logger"
	_ "net/http/pprof"
)

func main() {
	// 从consul获取配置
	conf, err := configstruct.GetConfig()
	if err != nil {
		logger.Error("service load config fail: ", err)
	}

	var confInfo configstruct.SysConfig
	if err = conf.Get("product").Scan(&confInfo); err != nil {
		logger.Error(err)
		return
	}

	if err := bootstrap.RunService(&confInfo); err != nil {
		logger.Error("failed to start service: ", err)
	}
}
