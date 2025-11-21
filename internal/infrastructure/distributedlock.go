package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"github.com/zhanshen02154/product/internal/config"
	"go-micro.dev/v4/logger"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"sync/atomic"
	"time"
)

// 分布式锁接口
type DistributedLock interface {
	Lock(ctx context.Context) (bool, error)
	UnLock(ctx context.Context) (bool, error)
	GetKey(ctx context.Context) string
}

// EtcdLock ETCD锁
type EtcdLock struct {
	session  *concurrency.Session
	mutex    *concurrency.Mutex
	prefix   string
	isLocked atomic.Bool
}

// 获取键名
func (l *EtcdLock) GetKey(ctx context.Context) string {
	return l.mutex.Key()
}

// 加锁
func (l *EtcdLock) Lock(ctx context.Context) (bool, error) {
	if l.isLocked.Load() {
		return false, errors.New(fmt.Sprintf("key: %s was locked", l.prefix))
	}
	l.mutex = concurrency.NewMutex(l.session, l.prefix)
	if err := l.mutex.Lock(ctx); err != nil {
		err = l.session.Close()
		if err != nil {
			return false, errors.New(fmt.Sprintf("prefix key: %s session close failed: %s", l.prefix, err))
		}
		return false, err
	}
	l.isLocked.Store(true)
	return true, nil
}

// 解锁
func (l *EtcdLock) UnLock(ctx context.Context) (bool, error) {
	defer func() {
		err := l.session.Close()
		if err != nil {
			logger.Fatalf(fmt.Sprintf("prefix key: %s session close failed: %s", l.prefix, err))
		}
	}()
	if err := l.mutex.Unlock(ctx); err != nil {
		return false, err
	}
	l.isLocked.Store(false)
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
	session, err := concurrency.NewSession(elm.ecli, concurrency.WithTTL(ttl), concurrency.WithContext(ctx))
	if err != nil {
		logger.Infof("failed to create session: %v", err)
		err = session.Close()
		if err != nil {
			return nil, err
		}
		return nil, err
	}
	return &EtcdLock{
		session: session,
		prefix:  fmt.Sprintf("%slock/%s", elm.prefix, key),
	}, nil
}

// NewEtcdLockManager 创建分布式锁
func NewEtcdLockManager(conf *config.Etcd) (LockManager, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:        conf.Hosts,
		AutoSyncInterval: time.Duration(conf.AutoSyncInterval) * time.Second,
		DialTimeout:      time.Duration(conf.DialTimeout) * time.Second,
		Username:         conf.Username,
		Password:         conf.Password,
	})
	if err != nil {
		return nil, err
	}
	logger.Info("ETCD was stared")
	return &EtcdLockManager{ecli: client, prefix: conf.Prefix}, nil
}
