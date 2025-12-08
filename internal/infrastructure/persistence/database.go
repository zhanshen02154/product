package persistence

import (
	"fmt"
	"github.com/zhanshen02154/product/internal/config"
	"go-micro.dev/v4/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/url"
	"time"
)

// InitDB 加载数据库
func InitDB(confInfo *config.MySqlConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s",
		confInfo.User,
		confInfo.Password,
		confInfo.Host,
		confInfo.Port,
		confInfo.Database,
		confInfo.Charset,
		url.QueryEscape(confInfo.Loc),
	)
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		SkipInitializeWithVersion: false,
		DefaultStringSize:         255,
	}), &gorm.Config{SkipDefaultTransaction: true, PrepareStmt: true})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if sqlDB == nil {
		return nil, fmt.Errorf("获取SQL DB失败: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(confInfo.MaxOpenConns)
	sqlDB.SetMaxIdleConns(confInfo.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(confInfo.ConnMaxLifeTime) * time.Second)

	// 验证连接
	if err = sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接验证失败: %w", err)
	}

	logger.Info("数据库连接成功")
	return db, nil
}
