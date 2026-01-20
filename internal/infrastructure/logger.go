package infrastructure

import (
	"context"
	"errors"
	metadatahelper "github.com/zhanshen02154/product/pkg/metadata"
	"go-micro.dev/v4/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"strings"
	"sync"
	"time"
)

type LogWrapper struct {
	logger            *zap.Logger
	level             zapcore.Level
	requestSlowTime   int64
	subscribeSlowTime int64
	striBuilderPool   sync.Pool
}

type Option func(p *LogWrapper)

// RequestLogWrapper 请求日志
func (w *LogWrapper) RequestLogWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		startTime := time.Now()
		err := fn(ctx, req, rsp)
		duration := time.Since(startTime).Milliseconds()
		baseFields := make([]zap.Field, 0, 12)
		baseFields = append(baseFields,
			zap.String("type", "request"),
			zap.String("trace_id", metadatahelper.GetTraceIdFromSpan(ctx)),
			zap.String("method", req.Method()),
			zap.String("endpoint", req.Endpoint()),
			zap.String("content_type", req.ContentType()),
			zap.String("user_agent", metadatahelper.GetValueFromMetadata(ctx, "user-agent")),
			zap.String("accept_encoding", metadatahelper.GetValueFromMetadata(ctx, "accept-encoding")),
			zap.String("remote", metadatahelper.GetValueFromMetadata(ctx, "Remote")),
			zap.Int64("duration", duration),
		)
		switch {
		case err != nil:
			w.logger.Error("Request failed: "+err.Error(), baseFields...)
		case duration > w.requestSlowTime && err == nil && duration > 0:
			w.logger.Warn("Slow request", baseFields...)
		default:
			w.logger.Info("Request processed", baseFields...)
		}
		return err
	}
}

// SubscribeWrapper 订阅事件记录日志
func (w *LogWrapper) SubscribeWrapper() server.SubscriberWrapper {
	return func(next server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, msg server.Message) error {
			startTime := time.Now()
			err := next(ctx, msg)
			duration := time.Since(startTime).Milliseconds()
			baseFields := make([]zap.Field, 0, 12)
			baseFields = append(baseFields, zap.String("type", "subscribe"),
				zap.String("trace_id", metadatahelper.GetTraceIdFromSpan(ctx)),
				zap.String("event_id", metadatahelper.GetValueFromMetadata(ctx, "Event_id")),
				zap.String("topic", msg.Topic()),
				zap.String("source", metadatahelper.GetValueFromMetadata(ctx, "Source")),
				zap.String("schema_version", metadatahelper.GetValueFromMetadata(ctx, "Schema_version")),
				zap.String("grpc_accept_encoding", metadatahelper.GetValueFromMetadata(ctx, "Grpc-Accept-Encoding")),
				zap.String("remote", metadatahelper.GetValueFromMetadata(ctx, "Remote")),
				zap.String("accept_encoding", metadatahelper.GetValueFromMetadata(ctx, "Accept-Encoding")),
				zap.String("key", metadatahelper.GetValueFromMetadata(ctx, "Pkey")),
				zap.Int64("duration", duration))
			switch {
			case err != nil:
				w.logger.Error("Event subscribe handler failed: "+err.Error(), baseFields...)
			case duration > w.requestSlowTime && duration > 0:
				w.logger.Warn("Event subscribe slow", baseFields...)
			default:
				w.logger.Info("Event subscribe handler processed", baseFields...)
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
	l.level = level
	return l
}

// Info Info日志
func (l *gormLogger) Info(ctx context.Context, str string, args ...interface{}) {
	if l.level < logger.Info {
		return
	}
	l.logger.Sugar().Infof(str, args...)
}

// Warn Warn日志
func (l *gormLogger) Warn(ctx context.Context, str string, args ...interface{}) {
	if l.level < logger.Warn {
		return
	}
	l.logger.Sugar().Warnf(str, args...)
}

// Error日志
func (l *gormLogger) Error(ctx context.Context, str string, args ...interface{}) {
	if l.level < logger.Error {
		return
	}
	l.logger.Sugar().Errorf(str, args...)
}

// Trace Trace日志
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// 获取运行时间
	elapsed := time.Since(begin).Milliseconds()

	// Gorm 错误
	switch {
	case err != nil && l.level >= logger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		sql, rows := fc()
		l.logger.Error("database query error: "+err.Error(),
			zap.String("type", "sql"),
			zap.String("trace_id", metadatahelper.GetTraceIdFromSpan(ctx)),
			zap.String("sql", sql),
			zap.Int64("time", elapsed),
			zap.Int64("rows", rows),
		)
	case l.slowThreshold != 0 && elapsed > l.slowThreshold && l.level >= logger.Warn:
		sql, rows := fc()
		l.logger.Warn("database query slow",
			zap.String("type", "sql"),
			zap.String("trace_id", metadatahelper.GetTraceIdFromSpan(ctx)),
			zap.String("sql", sql),
			zap.Int64("time", elapsed),
			zap.Int64("rows", rows),
		)
	case l.level >= logger.Info:
		sql, rows := fc()
		l.logger.Info("database query info",
			zap.String("type", "sql"),
			zap.String("trace_id", metadatahelper.GetTraceIdFromSpan(ctx)),
			zap.String("sql", sql),
			zap.Int64("time", elapsed),
			zap.Int64("rows", rows),
		)
	}
}

// NewGromLogger 创建GORM Logger
func NewGromLogger(zapLogger *zap.Logger, level zapcore.Level) logger.Interface {
	gormLevel := logger.Info
	switch level {
	case zap.WarnLevel:
		gormLevel = logger.Warn
		break
	case zap.ErrorLevel:
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
func NewLogWrapper(opts ...Option) *LogWrapper {
	w := LogWrapper{
		striBuilderPool: sync.Pool{New: func() interface{} {
			return &strings.Builder{}
		}},
	}
	for _, opt := range opts {
		opt(&w)
	}
	return &w
}

// WithRequestSlowThreshold 慢请求时间
func WithRequestSlowThreshold(timeout int64) Option {
	return func(p *LogWrapper) {
		p.requestSlowTime = timeout
	}
}

// WithSubscribeSlowThreshold 订阅事件处理延迟时间
func WithSubscribeSlowThreshold(timeout int64) Option {
	return func(p *LogWrapper) {
		p.subscribeSlowTime = timeout
	}
}

// WithZapLogger 设置Logger
func WithZapLogger(zapLogger *zap.Logger) Option {
	return func(p *LogWrapper) {
		p.logger = zapLogger
	}
}

// FindZapLogLevel zap日志级别
func FindZapLogLevel(level string) zapcore.Level {
	zapLevel := zap.DebugLevel
	switch level {
	case "info":
		zapLevel = zap.InfoLevel
		break
	case "warn":
		zapLevel = zap.WarnLevel
		break
	case "error":
		zapLevel = zap.ErrorLevel
		break
	case "fatal":
		zapLevel = zap.FatalLevel
	case "panic":
		zapLevel = zap.DPanicLevel
		break
	}
	return zapLevel
}

// FindZapAtomicLogLevel 获取原子日志级别
func FindZapAtomicLogLevel(level string) zap.AtomicLevel {
	var atomicLevel zap.AtomicLevel
	// 将字符串形式的日志级别（如从consul获取的"debug"）转换为zap.AtomicLevel
	switch level {
	case "debug":
		atomicLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		atomicLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		atomicLevel = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		atomicLevel = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		atomicLevel = zap.NewAtomicLevelAt(zap.InfoLevel) // 默认级别
	}
	return atomicLevel
}
