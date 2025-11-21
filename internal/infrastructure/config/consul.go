package config

import (
	"fmt"
	"github.com/go-micro/plugins/v4/config/source/consul"
	config2 "github.com/zhanshen02154/product/internal/config"
	"go-micro.dev/v4/config"
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
