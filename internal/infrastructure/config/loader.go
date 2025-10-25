package config

import (
	"fmt"
	"git.imooc.com/zhanshen1614/product/internal/config"
	config2 "github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/config/encoder/yaml"
	"github.com/micro/go-micro/v2/config/source"
	"github.com/micro/go-plugins/config/source/consul/v2"
	"os"
)

// LoadSystemConfig 获取系统配置
func LoadSystemConfig() (*config.SysConfig, error) {
	consulHost := getEnv("CONSUL_HOST", "192.168.83.131")
	consulPort := getEnv("CONSUL_PORT", "8500")
	consulPrefix := getEnv("CONSUL_PREFIX", "product")
	consulSource := consul.NewSource(
		// Set configuration address
		consul.WithAddress(fmt.Sprintf("%s:%s", consulHost, consulPort)),
		//前缀 默认：/micro/product
		consul.WithPrefix(consulPrefix),
		//consul.StripPrefix(true),
		source.WithEncoder(yaml.NewEncoder()),
	)
	configInfo, err := config2.NewConfig()
	if err != nil {
		return nil, err
	}

	err = configInfo.Load(consulSource)
	if err != nil {
		return nil, err
	}

	var sysConfig config.SysConfig
	if err = configInfo.Get(consulPrefix).Scan(&sysConfig); err != nil {
		return nil, err
	}

	return &sysConfig, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
