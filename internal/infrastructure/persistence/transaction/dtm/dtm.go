package dtm

import (
	"context"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/google/uuid"
)

type Server struct {
	grpcServer string
}

// NewServer 新建DTM服务器
func NewServer(host string) *Server {
	dtmcli.SetBarrierTableName("products.barrier")
	return &Server{grpcServer: host}
}

// BeginGrpcSaga 启动GrpcSaga事务
func (dtmSrv *Server) BeginGrpcSaga(ctx context.Context) *dtmgrpc.SagaGrpc {
	return dtmgrpc.NewSagaGrpc(dtmSrv.grpcServer, uuid.New().String())
}
