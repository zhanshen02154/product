package infrastructure

import (
	"context"
	"fmt"
	"github.com/zhanshen02154/product/internal/config"
	"go-micro.dev/v4/logger"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"sync"
	"time"
)

// DistributedLock 分布式锁接口
type DistributedLock interface {
	Lock(ctx context.Context) error
	TryLock(ctx context.Context) error
	UnLock(ctx context.Context) error
	GetKey(ctx context.Context) string
}

// EtcdLock ETCD锁
type EtcdLock struct {
	mutex   *concurrency.Mutex
	session *concurrency.Session
}

// GetKey 获取键名
func (l *EtcdLock) GetKey(ctx context.Context) string {
	return l.mutex.Key()
}

// Lock 加锁
func (l *EtcdLock) Lock(ctx context.Context) error {
	return l.mutex.Lock(ctx)
}

// TryLock 加锁（尝试获取锁）
func (l *EtcdLock) TryLock(ctx context.Context) error {
	return l.mutex.TryLock(ctx)
}

// UnLock 解锁
func (l *EtcdLock) UnLock(ctx context.Context) error {
	timeoutCtx, ctxCancelFunc := context.WithTimeout(context.Background(), time.Second*3)
	defer ctxCancelFunc()
	err := l.mutex.Unlock(timeoutCtx)
	closeErr := l.session.Close()
	if err != nil {
		return fmt.Errorf("failed to unlock %s: %v", l.mutex.Key(), err)
	}
	if closeErr != nil {
		logger.Error("failed to close session: ", err)
	}
	return nil
}

// LockManager 分布式锁管理器
type LockManager interface {
	NewLock(ctx context.Context, key string, ttl int) (DistributedLock, error)
	Close() error
}

// EtcdLockManager ETCD分布式锁
type EtcdLockManager struct {
	ecli     *clientv3.Client
	prefix   string
	isClosed bool
	mu       sync.RWMutex
}

// Close 关闭客户端
func (elm *EtcdLockManager) Close() error {
	elm.mu.Lock()
	defer elm.mu.Unlock()

	elm.isClosed = true
	return elm.ecli.Close()
}

// NewLock 创建锁
func (elm *EtcdLockManager) NewLock(ctx context.Context, key string, ttl int) (DistributedLock, error) {
	elm.mu.RLock()
	defer elm.mu.RUnlock()

	if elm.isClosed {
		return nil, fmt.Errorf("etcd client was closed")
	}
	session, err := concurrency.NewSession(elm.ecli, concurrency.WithTTL(ttl))
	if err != nil {
		return nil, err
	}
	mutex := concurrency.NewMutex(session, fmt.Sprintf("%slock/%s", elm.prefix, key))
	return &EtcdLock{
		mutex:   mutex,
		session: session,
	}, nil
}

// NewEtcdLockManager 创建分布式锁管理器
func NewEtcdLockManager(conf *config.Etcd) (LockManager, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:            conf.Hosts,
		DialTimeout:          10 * time.Second,
		Username:             conf.Username,
		Password:             conf.Password,
		DialKeepAliveTime:    10 * time.Second,
		DialKeepAliveTimeout: 5 * time.Second,
		MaxCallRecvMsgSize:   10 * 1024 * 1024,
		MaxCallSendMsgSize:   10 * 1024 * 1024,
	})
	if err != nil {
		return nil, err
	}
	logger.Info("ETCD was stared")
	return &EtcdLockManager{ecli: client, prefix: conf.Prefix}, nil
}
