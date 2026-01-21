package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-redsync/redsync/v4"
	redsyncredis "github.com/go-redsync/redsync/v4/redis"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"github.com/zhanshen02154/product/internal/config"
)

// DistributedLock 分布式锁接口
type DistributedLock interface {
	Lock(ctx context.Context) error
	TryLock(ctx context.Context) error
	UnLock(ctx context.Context) error
	GetKey() string
}

// LockManager 分布式锁管理器
type LockManager interface {
	NewLock(key string, ttl int) DistributedLock
	CheckHealth() error
	Close() error
}

type redisLockManager struct {
	client     *redis.Client
	rs         *redsync.Redsync
	prefix     string
	tries      int
	retryDelay time.Duration
	pool       redsyncredis.Pool
	isClosed   bool
	mu         sync.Mutex
}

type redisLock struct {
	m *redsync.Mutex
}

// NewRedisLockManager 创建Redis分布式锁
func NewRedisLockManager(conf *config.Redis) (LockManager, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         conf.Addr,
		Password:     conf.Password,
		DB:           conf.LockDB,
		PoolSize:     conf.PoolSize,
		DialTimeout:  time.Duration(conf.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(conf.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(conf.WriteTimeout) * time.Second,
		MinIdleConns: conf.MinIdleConns,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	lm := &redisLockManager{
		prefix:     conf.Prefix + ":lock:",
		tries:      conf.LockTries,
		retryDelay: time.Duration(conf.LockRetryDelay) * time.Millisecond,
		pool:       goredis.NewPool(client),
		client:     client,
		isClosed:   false,
		mu:         sync.Mutex{},
	}
	lm.rs = redsync.New(lm.pool)

	return lm, nil
}

// Close 关闭
func (rlm *redisLockManager) Close() error {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	// 检查是否已关闭
	if rlm.isClosed {
		return errors.New("redis lock manager already closed")
	}

	rlm.isClosed = true
	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer ctxCancel()
	var errs []error
	if rlm.client != nil {
		if err := rlm.client.Close(); err != nil {
			errs = append(errs, err)
		}
		rlm.client = nil
	}

	if rlm.pool != nil {
		if p, pErr := rlm.pool.Get(ctx); pErr != nil {
			errs = append(errs, pErr)
		} else {
			if err := p.Close(); err != nil {
				errs = append(errs, err)
			}
		}
		rlm.pool = nil
	}

	rlm.rs = nil

	// 综合所有错误并返回
	if len(errs) > 0 {
		// 多个错误合并为一个错误返回
		var combinedErr error
		for _, err := range errs {
			if combinedErr == nil {
				combinedErr = err
			} else {
				combinedErr = fmt.Errorf("%v; %w", combinedErr, err)
			}
		}
		return combinedErr
	}
	return nil
}

// NewLock 创建锁
func (rlm *redisLockManager) NewLock(key string, ttl int) DistributedLock {
	return &redisLock{m: rlm.rs.NewMutex(rlm.prefix+key,
		redsync.WithTries(rlm.tries),
		redsync.WithRetryDelay(rlm.retryDelay),
		redsync.WithExpiry(time.Second*time.Duration(ttl)),
	)}
}

// CheckHealth 健康检查
func (rlm *redisLockManager) CheckHealth() error {
	return rlm.client.Ping(context.Background()).Err()
}

// TryLock 非阻塞式加锁
func (rl *redisLock) TryLock(ctx context.Context) error {
	return rl.m.TryLockContext(ctx)
}

// UnLock 解锁
func (rl *redisLock) UnLock(ctx context.Context) error {
	_, err := rl.m.UnlockContext(ctx)
	return err
}

// GetKey 获取Key
func (rl *redisLock) GetKey() string {
	return rl.m.Name()
}

// Lock 阻塞式加锁
func (rl *redisLock) Lock(ctx context.Context) error {
	return rl.m.LockContext(ctx)
}
