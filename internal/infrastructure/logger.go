package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	metadatahelper "github.com/zhanshen02154/product/pkg/metadata"
	micrologger "go-micro.dev/v4/logger"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/server"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"strconv"
	"strings"
	"time"
)

type LogWrapper struct {
	logger *zap.Logger
}

// RequestLogWrapper 请求日志
func (w *LogWrapper) RequestLogWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		traceId := metadatahelper.GetValueFromMetadata(ctx, "Trace_id")
		if traceId == "" {
			traceId = uuid.New().String()
		}
		traceCtx := metadata.Set(ctx, "Trace_id", traceId)
		startTime := time.Now()
		err := fn(traceCtx, req, rsp)
		duration := time.Since(startTime).Milliseconds()
		userAgent := metadatahelper.GetValueFromMetadata(ctx, "user-agent")
		remote := metadatahelper.GetValueFromMetadata(ctx, "Remote")
		acceptEncoding := metadatahelper.GetValueFromMetadata(ctx, "accept-encoding")
		logFields := []zap.Field{
			zap.String("type", "request"),
			zap.String("trace_id", traceId),
			zap.String("service", req.Service()),
			zap.String("method", req.Method()),
			zap.String("endpoint", req.Endpoint()),
			zap.String("content_type", req.ContentType()),
			zap.String("user_agent", userAgent),
			zap.String("accept_encoding", acceptEncoding),
			zap.String("remote", remote),
			zap.Int64("duration", duration),
		}
		if err != nil {
			w.logger.Error(fmt.Sprintf("request failed: %s", err.Error()), logFields...)
		} else {
			w.logger.Info("request success", logFields...)
		}
		return err
	}
}

// SubscribeWrapper 订阅事件记录日志
func (w *LogWrapper) SubscribeWrapper() server.SubscriberWrapper {
	return func(next server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, msg server.Message) error {
			var err error
			startTime := time.Now()
			err = next(ctx, msg)
			duration := time.Since(startTime).Milliseconds()
			var strBuilder strings.Builder
			if err != nil {
				strBuilder.WriteString(fmt.Sprintf("failed to subscribe on %s, error: %s", msg.Topic(), err.Error()))
			} else {
				strBuilder.WriteString(fmt.Sprintf("topic: %s handle success", msg.Topic()))
			}
			publishedAt, err := strconv.ParseInt(metadatahelper.GetValueFromMetadata(ctx, "Timestamp"), 10, 64)
			if err != nil {
				micrologger.Error("failed to convert publushed at: ", err.Error())
			}
			logFields := []zap.Field{
				zap.String("type", "subscribe"),
				zap.String("trace_id", metadatahelper.GetTraceIdFromSpan(ctx)),
				zap.String("event_id", metadatahelper.GetValueFromMetadata(ctx, "Event_id")),
				zap.String("topic", msg.Topic()),
				zap.String("source", metadatahelper.GetValueFromMetadata(ctx, "Source")),
				zap.String("schema_version", metadatahelper.GetValueFromMetadata(ctx, "Schema_version")),
				zap.Int64("published_at", publishedAt),
				zap.String("grpc_accept_encoding", metadatahelper.GetValueFromMetadata(ctx, "Grpc-Accept-Encoding")),
				zap.String("remote", metadatahelper.GetValueFromMetadata(ctx, "Remote")),
				zap.String("accept_encoding", metadatahelper.GetValueFromMetadata(ctx, "Accept-Encoding")),
				zap.String("key", metadatahelper.GetValueFromMetadata(ctx, "Pkey")),
				zap.Int64("duration", duration),
			}
			if err != nil {
				w.logger.Error(strBuilder.String(), logFields...)
			} else {
				w.logger.Info(strBuilder.String(), logFields...)
			}
			return err
		}
	}
}

// GORM Logger
type gormLogger struct {
	logger        *zap.Logger
	slowThreshold int64
	level         logger.LogLevel
}

func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

// Info Info日志
func (l *gormLogger) Info(ctx context.Context, str string, args ...interface{}) {
	l.logger.Sugar().Debugf(str, args...)
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
	if l.level < logger.Info {
		return
	}
	// 获取运行时间
	elapsed := time.Since(begin).Milliseconds()
	// 获取 SQL 请求和返回条数
	sql, rows := fc()
	// 通用字段
	traceId := metadatahelper.GetTraceIdFromSpan(ctx)
	logFields := []zap.Field{
		zap.String("type", "sql"),
		zap.String("trace_id", traceId),
		zap.String("sql", sql),
		zap.Int64("time", elapsed),
		zap.Int64("rows", rows),
	}

	// Gorm 错误
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// 其他错误使用 error 等级
		logFields = append(logFields, zap.Error(err))
		l.logger.Error("Database Error", logFields...)
	}

	// 慢查询日志
	if l.slowThreshold != 0 && elapsed > l.slowThreshold {
		l.logger.Warn("Database Slow Log", logFields...)
	}

	// 记录所有 SQL 请求
	if l.level == logger.Info {
		l.logger.Info("Database Query", logFields...)
	}
}

// NewGromLogger 创建GORM Logger
func NewGromLogger(zapLogger *zap.Logger, level int) logger.Interface {
	gormLevel := logger.Info
	switch level {
	case 1:
		gormLevel = logger.Info
		break
	case 2:
		gormLevel = logger.Warn
		break
	case 3:
		gormLevel = logger.Error
		break
	}
	return &gormLogger{
		logger:        zapLogger,
		slowThreshold: 200,
		level:         gormLevel,
	}
}

// NewLogWrapper 创建日志包装器
func NewLogWrapper(logger *zap.Logger) LogWrapper {
	return LogWrapper{logger: logger}
}
