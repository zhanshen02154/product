package infrastructure

import (
	"github.com/go-micro/plugins/v4/registry/consul"
	"github.com/hashicorp/consul/api"
	"github.com/zhanshen02154/product/internal/config"
	"go-micro.dev/v4/registry"
	"time"
)

// ConsulRegister consul注册
func ConsulRegister(confInfo *config.ConsulInfo) registry.Registry {
	queryOpts := &api.QueryOptions{
		WaitTime: time.Duration(60) * time.Second,
	}
	return consul.NewRegistry(
		registry.Addrs(confInfo.RegistryAddrs...),
		registry.Timeout(time.Duration(confInfo.Timeout)*time.Second),
		consul.QueryOptions(queryOpts),
		consul.Config(&api.Config{
			WaitTime: time.Duration(60) * time.Second,
		}),
	)
}
