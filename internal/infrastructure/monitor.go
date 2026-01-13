package infrastructure

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go-micro.dev/v4/logger"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync"
	"sync/atomic"
)

// MonitorServer pprof服务器
type MonitorServer struct {
	pprof        *http.Server
	prom         *http.Server
	wg           sync.WaitGroup
	shutdownFlag atomic.Bool
}

// NewMonitorServer 创建监控服务器
func NewMonitorServer(addr string) *MonitorServer {
	runtime.SetBlockProfileRate(10000)
	runtime.SetCPUProfileRate(100)
	runtime.SetMutexProfileFraction(1)
	return &MonitorServer{
		pprof:        &http.Server{Addr: addr},
		prom:         &http.Server{Addr: ":9092"},
		wg:           sync.WaitGroup{},
		shutdownFlag: atomic.Bool{},
	}
}

// Start 启动pprof
func (srv *MonitorServer) Start() {
	srv.wg.Add(2)
	logger.Info("Starting pprof servers")
	go func() {
		defer srv.wg.Done()
		err := srv.pprof.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Error("Faild to start pprof server: " + err.Error())
			return
		}
	}()
	srv.startMetrics()
}

func (srv *MonitorServer) startMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		defer srv.wg.Done()
		err := srv.prom.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start metrics server: " + err.Error())
			return
		}
	}()
}

// Close 关闭pprof
func (srv *MonitorServer) Close(ctx context.Context) error {
	srv.shutdownFlag.Store(true)
	if err := srv.pprof.Shutdown(ctx); err != nil {
		logger.Error("Failed to close pprof server: " + err.Error())
	}
	if err := srv.prom.Shutdown(ctx); err != nil {
		logger.Error("Failed to close prometheus mertrics server: " + err.Error())
	}
	srv.wg.Wait()
	logger.Info("Monitor server was closed")
	return nil
}
