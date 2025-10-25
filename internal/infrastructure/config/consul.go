package config

import (
	"fmt"
	config2 "git.imooc.com/zhanshen1614/product/internal/config"
	"github.com/micro/go-micro/v2/config"
	"github.com/micro/go-plugins/config/source/consul/v2"
)

func GetConsulConfig(conf *config2.SysConfig) (config.Config, error) {
	consulSource := consul.NewSource(
		// Set configuration address
		consul.WithAddress(fmt.Sprintf("%s:%d", conf.Consul.Addr, conf.Consul.Port)),
		//前缀 默认：/micro/product
		consul.WithPrefix(conf.Consul.Prefix),
	)
	configInfo, err := config.NewConfig()
	if err != nil {
		return nil, err
	}

	// Load config
	err = configInfo.Load(consulSource)
	return configInfo, err
}