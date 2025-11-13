package infrastructure

import (
	"context"
	"github.com/micro/go-micro/v2/util/log"
	"gorm.io/gorm"
	"net/http"
	"sync"
	"sync/atomic"
)

// ProbeServer 健康检查探针
type ProbeServer struct {
	server         *http.Server
	db             *gorm.DB
	wg             sync.WaitGroup
	isShuttingDown atomic.Bool
}

func NewProbeServer(port string, db *gorm.DB) *ProbeServer {
	mx := http.NewServeMux()
	mx.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("OK"))
	})
	mx.HandleFunc("/ready", func(writer http.ResponseWriter, request *http.Request) {
		sqlDB, err := db.DB()
		if err != nil {
			writer.WriteHeader(http.StatusServiceUnavailable)
			writer.Write([]byte("Not Ready"))
		} else {
			if err = sqlDB.Ping(); err != nil {
				writer.WriteHeader(http.StatusServiceUnavailable)
				writer.Write([]byte("Not Ready"))
			} else {
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte("OK"))
			}
		}
	})
	return &ProbeServer{
		server: &http.Server{Addr: port, Handler: mx},
		db:     db,
	}
}

// Start 启动服务器
func (probeServe *ProbeServer) Start() error {
	probeServe.wg.Add(1)
	go func() {
		defer probeServe.wg.Done()
		err := probeServe.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Health check server error: %v", err)
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
	log.Info("健康检查服务器已关闭")
	return nil
}
