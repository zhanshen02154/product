package server

import (
	"context"
	"go-micro.dev/v4/logger"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
)

// PprofServer pprof服务器
type PprofServer struct {
	server       *http.Server
	wg           sync.WaitGroup
	shutdownFlag atomic.Bool
}

// NewPprofServer 创建pprof服务器
func NewPprofServer(addr string) *PprofServer {
	runtime.SetBlockProfileRate(1)
	runtime.SetCPUProfileRate(1)
	runtime.SetMutexProfileFraction(1)
	return &PprofServer{
		server:       &http.Server{Addr: addr},
		wg:           sync.WaitGroup{},
		shutdownFlag: atomic.Bool{},
	}
}

// Start 启动pprof
func (srv *PprofServer) Start() {
	var wg sync.WaitGroup
	wg.Add(1)
	logger.Info("启动pprof")
	go func() {
		defer wg.Done()
		err := srv.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Error("pprof服务器启动失败")
			return
		}
	}()
}

// Close 关闭pprof
func (srv *PprofServer) Close(ctx context.Context) error {
	srv.shutdownFlag.Store(true)
	if err := srv.server.Shutdown(ctx); err != nil {
		return err
	}
	srv.wg.Wait()
	logger.Info("pprof服务已关闭")
	return nil
}
