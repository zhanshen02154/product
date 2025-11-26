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
	Lock(ctx context.Context) (bool, error)
	TryLock(ctx context.Context) (bool, error)
	UnLock(ctx context.Context) (bool, error)
	GetKey(ctx context.Context) string
}

// EtcdLock ETCD锁
type EtcdLock struct {
	mutex *concurrency.Mutex
}

// GetKey 获取键名
func (l *EtcdLock) GetKey(ctx context.Context) string {
	return l.mutex.Key()
}

// Lock 加锁
func (l *EtcdLock) Lock(ctx context.Context) (bool, error) {
	if err := l.mutex.Lock(ctx); err != nil {
		return false, err
	}
	return true, nil
}

// TryLock 加锁（尝试获取锁）
func (l *EtcdLock) TryLock(ctx context.Context) (bool, error) {
	if err := l.mutex.TryLock(ctx); err != nil {
		return false, err
	}
	return true, nil
}

// UnLock 解锁
func (l *EtcdLock) UnLock(ctx context.Context) (bool, error) {
	if err := l.mutex.Unlock(ctx); err != nil {
		return false, err
	}
	return true, nil
}

// LockManager 分布式锁管理器
type LockManager interface {
	NewLock(ctx context.Context, key string) (DistributedLock, error)
	Close() error
}

// EtcdLockManager ETCD分布式锁
type EtcdLockManager struct {
	ecli     *clientv3.Client
	prefix   string
	session  *concurrency.Session
	isClosed bool
	mu       sync.RWMutex
}

// Close 关闭客户端
func (elm *EtcdLockManager) Close() error {
	elm.mu.Lock()
	defer elm.mu.Unlock()

	elm.isClosed = true
	if elm.session != nil {
		err := elm.session.Close()
		if err != nil {
			logger.Errorf("failed to close etcd session: ", err)
		}
	}
	return elm.ecli.Close()
}

// NewLock 创建锁
func (elm *EtcdLockManager) NewLock(ctx context.Context, key string) (DistributedLock, error) {
	elm.mu.RLock()
	defer elm.mu.RUnlock()

	if elm.isClosed {
		return nil, fmt.Errorf("etcd client was closed")
	}
	mutex := concurrency.NewMutex(elm.session, fmt.Sprintf("%slock/%s", elm.prefix, key))
	return &EtcdLock{
		mutex: mutex,
	}, nil
}

// NewEtcdLockManager 创建分布式锁
func NewEtcdLockManager(conf *config.Etcd) (LockManager, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: conf.Hosts,
		//AutoSyncInterval: time.Duration(conf.AutoSyncInterval) * time.Second,
		DialTimeout: time.Duration(conf.DialTimeout) * time.Second,
		Username:    conf.Username,
		Password:    conf.Password,
	})
	if err != nil {
		return nil, err
	}
	session, err := concurrency.NewSession(client, concurrency.WithTTL(30))
	if err != nil {
		client.Close()
		return nil, err
	}
	logger.Info("ETCD was stared")
	return &EtcdLockManager{ecli: client, prefix: conf.Prefix, session: session}, nil
}
