package infrastructure

import (
	"context"
	"github.com/micro/go-micro/v2/util/log"
	"gorm.io/gorm"
	"net/http"
	"sync"
	"time"
)

// ProbeServer 健康检查探针
type ProbeServer struct {
	server *http.Server
	db     *gorm.DB
	wg     sync.WaitGroup
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
			if err := sqlDB.Ping(); err != nil {
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
		wg:     sync.WaitGroup{},
	}
}

// Start 启动服务器
func (probeServe *ProbeServer) Start(ctx context.Context) {
	probeServe.wg.Add(1)
	go func() {
		defer probeServe.wg.Done()
		go func() {
			err := probeServe.server.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				log.Fatalf("Health check server start error: %v", err)
			}
		}()

		<-ctx.Done()
		log.Info("收到关闭信号，正在停止健康检查服务器...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := probeServe.server.Shutdown(shutdownCtx); err != nil {
			log.Infof("健康检查探针服务关闭错误: %v", err)
		} else {
			log.Info("健康检查探针服务已关闭")
		}
	}()
}

func (probeServe *ProbeServer) Wait() {
	probeServe.wg.Wait()
}
