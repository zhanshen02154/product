package registry

import (
	"context"
	"fmt"
	"git.imooc.com/zhanshen1614/product/internal/config"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-plugins/registry/consul/v2"
	"time"
)

// ConsulRegister consul注册
func ConsulRegister(confInfo *config.ConsulInfo) registry.Registry {
	return consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{
			fmt.Sprintf("%s:%d", confInfo.Addr, confInfo.Port),
		}
		options.Timeout = time.Duration(confInfo.Timeout) * time.Second
		options.Context = context.WithValue(context.Background(), "api.token", confInfo.Token)
	})
}
