package bootstrap

import (
	grpcserver "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/go-micro/plugins/v4/transport/grpc"
	"github.com/zhanshen02154/product/internal/config"
	"github.com/zhanshen02154/product/internal/infrastructure"
	"github.com/zhanshen02154/product/pkg/codec"
	"go-micro.dev/v4"
	"go-micro.dev/v4/server"
	"time"
)

func newServer(conf *config.SysConfig) micro.Option {
	// 注册到Consul
	consulRegistry := infrastructure.ConsulRegister(conf.Consul)
	return micro.Server(
		grpcserver.NewServer(
			server.Name(conf.Service.Name),
			server.Version(conf.Service.Version),
			server.Address(conf.Service.Listen),
			server.Transport(grpc.NewTransport()),
			server.Registry(consulRegistry),
			server.RegisterTTL(time.Duration(conf.Consul.RegisterTtl)*time.Second),
			server.RegisterInterval(time.Duration(conf.Consul.RegisterInterval)*time.Second),
			grpcserver.Codec("application/grpc+dtm_raw", codec.NewDtmCodec()),
		))
}
