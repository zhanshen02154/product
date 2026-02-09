package gorm

import (
	"context"
	"errors"
	"fmt"
	metadatahelper "github.com/zhanshen02154/product/pkg/metadata"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

// GORM Logger
type gormLogger struct {
	logger *zap.Logger
	logger.Config
}

func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.LogLevel = level
	return l
}

// Info Info日志
func (l *gormLogger) Info(ctx context.Context, str string, args ...interface{}) {
	l.logger.Sugar().Infof(str, args...)
}

// Warn Warn日志
func (l *gormLogger) Warn(ctx context.Context, str string, args ...interface{}) {
	l.logger.Sugar().Warnf(str, args...)
}

// Error日志
func (l *gormLogger) Error(ctx context.Context, str string, args ...interface{}) {
	l.logger.Sugar().Errorf(str, args...)
}

// Trace Trace日志
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	// 获取运行时间
	elapsed := time.Since(begin)

	// Gorm 错误
	switch {
	case err != nil && l.LogLevel >= logger.Error:
		sql, rows := fc()
		if errors.Is(err, gorm.ErrRecordNotFound) && !l.IgnoreRecordNotFoundError {
			l.logger.Warn("database query data not found: "+err.Error(),
				zap.String("type", "sql"),
				zap.String("trace_id", metadatahelper.GetTraceIdFromSpan(ctx)),
				zap.String("sql", sql),
				zap.Int64("time", elapsed.Milliseconds()),
				zap.Int64("rows", rows),
			)
		} else {
			l.logger.Error("database query error: "+err.Error(),
				zap.String("type", "sql"),
				zap.String("trace_id", metadatahelper.GetTraceIdFromSpan(ctx)),
				zap.String("sql", sql),
				zap.Int64("time", elapsed.Milliseconds()),
				zap.Int64("rows", rows),
			)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		l.logger.Warn(fmt.Sprintf("SLOW SQL >= %d ms", l.SlowThreshold.Milliseconds()),
			zap.String("type", "sql"),
			zap.String("trace_id", metadatahelper.GetTraceIdFromSpan(ctx)),
			zap.String("sql", sql),
			zap.Int64("time", elapsed.Milliseconds()),
			zap.Int64("rows", rows),
		)
	case l.LogLevel == logger.Info:
		sql, rows := fc()
		l.logger.Info("database query info",
			zap.String("type", "sql"),
			zap.String("trace_id", metadatahelper.GetTraceIdFromSpan(ctx)),
			zap.String("sql", sql),
			zap.Int64("time", elapsed.Milliseconds()),
			zap.Int64("rows", rows),
		)
	}
}

// NewGromLogger 创建GORM Logger
func NewGromLogger(zapLogger *zap.Logger, config logger.Config) logger.Interface {
	return &gormLogger{
		logger: zapLogger,
		Config: config,
	}
}

// GetLogLevel 获取GORM的日志级别
func GetLogLevel(level zapcore.Level) logger.LogLevel {
	switch level {
	case zap.WarnLevel:
		return logger.Warn
	case zap.ErrorLevel:
		return logger.Error
	}
	return logger.Info
}
