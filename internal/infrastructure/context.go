package infrastructure

import (
	"fmt"
	"github.com/zhanshen02154/product/internal/config"
	"github.com/zhanshen02154/product/internal/domain/repository"
	"github.com/zhanshen02154/product/internal/infrastructure/persistence"
	gorm2 "github.com/zhanshen02154/product/internal/infrastructure/persistence/gorm"
	"github.com/zhanshen02154/product/internal/infrastructure/persistence/transaction"
	"github.com/zhanshen02154/product/internal/infrastructure/persistence/transaction/dtm"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/registry"
	"gorm.io/gorm"
)

type ServiceContext struct {
	TxManager       transaction.TransactionManager
	LockManager     LockManager
	Conf            *config.SysConfig
	db              *gorm.DB
	OrderRepository repository.IProductRepository
	Dtm             *dtm.Server
}

func NewServiceContext(conf *config.SysConfig, serviceReg registry.Registry) (*ServiceContext, error) {
	db, err := persistence.InitDB(conf.Database)
	if err != nil {
		return nil, err
	}

	// 加载ETCD分布式锁
	lockMgr, err := NewEtcdLockManager(conf.Etcd)
	if err != nil {
		logger.Fatalf(fmt.Sprintf("failed to load lock manager: %v", err))
		return nil, err
	}
	return &ServiceContext{
		TxManager:       gorm2.NewGormTransactionManager(db),
		LockManager:     lockMgr,
		Conf:            conf,
		db:              db,
		OrderRepository: gorm2.NewProductRepository(db),
		Dtm:             dtm.NewServer(conf.Transaction.Host),
	}, nil
}

// Close 关闭所有服务
func (svc *ServiceContext) Close() {
	// 关闭数据库
	if err := svc.closeDB(); err != nil {
		logger.Fatalf("close database error: %v", err)
	}
	// 关闭ETCD
	if err := svc.closeEtcd(); err != nil {
		logger.Fatalf("close etcd error: %v", err)
	}
}

// 关闭数据库连接
func (svc *ServiceContext) closeDB() error {
	sqlDB, err := svc.db.DB()
	if err != nil {

		return err
	} else {
		logger.Info("Preparing to close GORM")
	}
	if err := sqlDB.Close(); err != nil {
		logger.Fatalf("Failed to close database instance: %v", err)
		return err
	} else {
		logger.Info("GORM数据库连接已关闭")
	}
	return nil
}

// 关闭ETCD
func (svc *ServiceContext) closeEtcd() error {
	err := svc.LockManager.Close()
	if err != nil {
		logger.Fatalf("Failed to close etcd lock manager: %v", err)
	} else {
		logger.Info("ETCD lock manager closed")
	}
	return err
}

// CheckHealth 检查是否健康
func (svc *ServiceContext) CheckHealth() error {
	sqlDB, err := svc.db.DB()
	if err != nil {
		return err
	}
	if err := sqlDB.Ping(); err != nil {
		logger.Fatalf("Failed to close database instance: %v", err)
	}
	return nil
}
