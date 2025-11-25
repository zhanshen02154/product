package infrastructure

import (
	"context"
	"fmt"
	"github.com/zhanshen02154/product/internal/config"
	"go-micro.dev/v4/logger"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
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
	session    *concurrency.Session
	mutex      *concurrency.Mutex
	cancelFunc context.CancelFunc
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
	defer func() {
		err := l.session.Close()
		if err != nil {
			logger.Errorf("prefix key: %s session close failed: %s", l.mutex.Key(), err)
		}
		if l.cancelFunc != nil {
			l.cancelFunc()
		}
	}()
	if err := l.mutex.Unlock(ctx); err != nil {
		return false, err
	}
	return true, nil
}

// LockManager 分布式锁管理器
type LockManager interface {
	NewLock(ctx context.Context, key string, ttl int) (DistributedLock, error)
	Close() error
}

// EtcdLockManager ETCD分布式锁
type EtcdLockManager struct {
	ecli   *clientv3.Client
	prefix string
}

// Close 关闭客户端
func (elm *EtcdLockManager) Close() error {
	return elm.ecli.Close()
}

// NewLock 创建锁
func (elm *EtcdLockManager) NewLock(ctx context.Context, key string, ttl int) (DistributedLock, error) {
	sessionCtx, sessionCanctlCtx := context.WithCancel(context.Background())
	session, err := concurrency.NewSession(elm.ecli, concurrency.WithTTL(ttl), concurrency.WithContext(sessionCtx))
	if err != nil {
		sessionCanctlCtx()
		return nil, err
	}
	mutex := concurrency.NewMutex(session, fmt.Sprintf("%slock/%s", elm.prefix, key))
	return &EtcdLock{
		session:    session,
		mutex:      mutex,
		cancelFunc: sessionCanctlCtx,
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
	logger.Info("ETCD was stared")
	return &EtcdLockManager{ecli: client, prefix: conf.Prefix}, nil
}
