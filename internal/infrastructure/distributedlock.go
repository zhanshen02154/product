package infrastructure

import (
	"context"
	"errors"
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

// ETCD锁
type etcdLock struct {
	mutex *concurrency.Mutex
}

// GetKey 获取键名
func (l *etcdLock) GetKey(ctx context.Context) string {
	return l.mutex.Key()
}

// Lock 加锁
func (l *etcdLock) Lock(ctx context.Context) error {
	return l.mutex.Lock(ctx)
}

// TryLock 加锁（尝试获取锁）
func (l *etcdLock) TryLock(ctx context.Context) error {
	return l.mutex.TryLock(ctx)
}

// UnLock 解锁
func (l *etcdLock) UnLock(ctx context.Context) error {
	timeoutCtx, ctxCancelFunc := context.WithTimeout(context.Background(), time.Second*3)
	defer ctxCancelFunc()
	if err := l.mutex.Unlock(timeoutCtx); err != nil {
		return fmt.Errorf("failed to unlock %s: %v", l.mutex.Key(), err)
	}
	return nil
}

// LockManager 分布式锁管理器
type LockManager interface {
	NewLock(ctx context.Context, key string, ttl int) (DistributedLock, error)
	Close() error
}

// etcdLockManager ETCD分布式锁
type etcdLockManager struct {
	ecli       *clientv3.Client
	prefix     string
	isClosed   bool
	mu         sync.RWMutex
	sessions   map[int]*concurrency.Session
	sessionsMu sync.Mutex
}

// Close 关闭客户端
func (elm *etcdLockManager) Close() error {
	elm.mu.Lock()
	defer elm.mu.Unlock()

	elm.isClosed = true
	elm.sessionsMu.Lock()
	for ttl, s := range elm.sessions {
		if s != nil {
			_ = s.Close()
		}
		delete(elm.sessions, ttl)
	}
	elm.sessions = nil
	elm.sessionsMu.Unlock()

	return elm.ecli.Close()
}

// NewLock 创建锁
func (elm *etcdLockManager) NewLock(ctx context.Context, key string, ttl int) (DistributedLock, error) {
	elm.mu.RLock()
	defer elm.mu.RUnlock()

	if elm.isClosed {
		return nil, errors.New("etcd client was closed")
	}
	session, err := elm.getOrCreateSession(ttl)
	if err != nil {
		return nil, err
	}
	mutex := concurrency.NewMutex(session, elm.prefix+key)
	return &etcdLock{
		mutex: mutex,
	}, nil
}

// getOrCreateSession 获取/创建会话
func (elm *etcdLockManager) getOrCreateSession(ttl int) (*concurrency.Session, error) {
	elm.sessionsMu.Lock()
	defer elm.sessionsMu.Unlock()

	if s, ok := elm.sessions[ttl]; ok && s != nil {
		return s, nil
	}
	s, err := concurrency.NewSession(elm.ecli, concurrency.WithTTL(ttl))
	if err != nil {
		return nil, err
	}
	if elm.sessions == nil {
		elm.sessions = make(map[int]*concurrency.Session)
	}
	elm.sessions[ttl] = s
	return s, nil
}

// NewEtcdLockManager 创建分布式锁
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
	return &etcdLockManager{ecli: client, prefix: conf.Prefix + "lock/", sessions: make(map[int]*concurrency.Session)}, nil
}
