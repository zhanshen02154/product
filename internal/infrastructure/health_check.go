package infrastructure

import (
	"context"
	"errors"
	"go-micro.dev/v4/logger"
	"net/http"
	"sync"
	"sync/atomic"
)

// ProbeServer 健康检查探针
type ProbeServer struct {
	server         *http.Server
	wg             sync.WaitGroup
	isShuttingDown atomic.Bool
	serviceContext *ServiceContext
}

func NewProbeServer(port string, serviceContext *ServiceContext) *ProbeServer {
	mx := http.NewServeMux()
	mx.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		err := serviceContext.CheckHealth()
		if err != nil {
			writer.WriteHeader(http.StatusServiceUnavailable)
			writer.Write([]byte("Not Ready"))
			logger.Error("health check failed: " + err.Error())
		} else {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("OK"))
			logger.Info("health check success")
		}
	})
	mx.HandleFunc("/ready", func(writer http.ResponseWriter, request *http.Request) {
		err := serviceContext.CheckHealth()
		if err != nil {
			writer.WriteHeader(http.StatusServiceUnavailable)
			writer.Write([]byte("Not Ready"))
			logger.Error("service was not ready: " + err.Error())
		} else {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("OK"))
			logger.Info("service was ready")
		}
	})
	return &ProbeServer{
		server:         &http.Server{Addr: port, Handler: mx},
		serviceContext: serviceContext,
	}
}

// Start 启动服务器
func (probeServe *ProbeServer) Start() error {
	probeServe.wg.Add(1)
	go func() {
		defer probeServe.wg.Done()
		err := probeServe.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Health check server error: " + err.Error())
		}
	}()
	return nil
}

// Shutdown 关闭服务器
func (probeServe *ProbeServer) Shutdown(ctx context.Context) error {
	probeServe.isShuttingDown.Store(true)
	if err := probeServe.server.Shutdown(ctx); err != nil {
		return err
	}
	probeServe.wg.Wait()
	logger.Info("Health check server shutdown success")
	return nil
}
