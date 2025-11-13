package registry

import (
	"github.com/hashicorp/consul/api"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-plugins/registry/consul/v2"
	"github.com/zhanshen02154/product/internal/config"
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
