package infrastructure

import (
	"errors"
	"github.com/zhanshen02154/product/internal/config"
	"go-micro.dev/v4/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
	"time"
)

// InitDB 加载数据库
func InitDB(confInfo *config.MySqlConfig, gLogger gormlogger.Interface) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       confInfo.Dsn,
		SkipInitializeWithVersion: false,
		DefaultStringSize:         255,
	}), &gorm.Config{SkipDefaultTransaction: true, PrepareStmt: true, Logger: gLogger})
	if err != nil {
		return nil, err
	}
	if err := db.Use(tracing.NewPlugin(tracing.WithoutMetrics())); err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if sqlDB == nil {
		return nil, errors.New("failed to get database connection")
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(confInfo.MaxOpenConns)
	sqlDB.SetMaxIdleConns(confInfo.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(confInfo.ConnMaxLifeTime) * time.Second)

	// 验证连接
	if err := sqlDB.Ping(); err != nil {
		return nil, errors.New("Cannot to connect database instance: " + err.Error())
	}

	logger.Info("Connect database success")
	return db, nil
}
