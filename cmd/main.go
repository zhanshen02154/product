package main

import (
	"context"
	microzap "github.com/go-micro/plugins/v4/logger/zap"
	"github.com/zhanshen02154/product/internal/bootstrap"
	configstruct "github.com/zhanshen02154/product/internal/config"
	"github.com/zhanshen02154/product/internal/infrastructure"
	"go-micro.dev/v4/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	_ "net/http/pprof"
	"time"
)

func main() {
	zapLogger, err := zap.NewProduction(
		zap.AddCallerSkip(2),
	)
	if err != nil {
		log.Panic("failed to start zap logger: ", err.Error())
	}
	defer zapLogger.Sync()

	consulSource, err := configstruct.GetConfig()
	if err != nil {
		zapLogger.Error("failed to load config: " + err.Error())
		return
	}

	var confInfo configstruct.SysConfig
	if err := consulSource.Get("order").Scan(&confInfo); err != nil {
		zapLogger.Error("failed convert config to struct: " + err.Error())
		return
	}
	// 检查配置
	if err := confInfo.CheckConfig(); err != nil {
		zapLogger.Error("failed to check config: " + err.Error())
		return
	}

	// 创建一个全新的zap logger
	loggerConfig := zap.Config{
		Level:            infrastructure.FindZapAtomicLogLevel(confInfo.Service.LogLevel),
		Development:      false,
		Encoding:         "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写日志级别
			EncodeTime:     zapcore.ISO8601TimeEncoder,    // 可读的时间格式[3](@ref)
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder, // 记录调用者信息[3](@ref)
		},
	}
	finalLogger, err := loggerConfig.Build(zap.Fields(
		zap.String("service", confInfo.Service.Name),
		zap.String("version", confInfo.Service.Version),
	), zap.AddCallerSkip(1))
	if err != nil {
		zapLogger.Error("Failed to build final logger", zap.Error(err))
		return
	}
	defer finalLogger.Sync()
	serverLogLevel := infrastructure.FindZapLogLevel(confInfo.Service.LogLevel)
	loggerMetadataMap := make(map[string]interface{})
	loggerMetadataMap["type"] = "core"
	microLogger, err := microzap.NewLogger(
		microzap.WithLogger(finalLogger),
	)
	if err != nil {
		zapLogger.Error("failed to load go micro logger: " + err.Error())
		return
	}
	logger.DefaultLogger = microLogger.Fields(loggerMetadataMap)

	// 链路追踪
	traceShutdown := initTracer(confInfo.Service.Name, confInfo.Service.Version, confInfo.Tracer)
	defer traceShutdown()

	gormLogger := infrastructure.NewGromLogger(finalLogger, serverLogLevel)
	serviceContext, err := infrastructure.NewServiceContext(&confInfo, gormLogger)
	if err != nil {
		logger.Error("error to load service context: ", err)
		return
	}

	if err := bootstrap.RunService(&confInfo, serviceContext, finalLogger); err != nil {
		logger.Error("failed to start service: ", err)
	}
}

// 加载OpenTelemetry链路追踪
func initTracer(serviceName string, version string, conf *configstruct.Tracer) func() {
	ctx := context.Background()
	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithTimeout(time.Duration(conf.Client.Timeout)*time.Second),
		otlptracegrpc.WithEndpoint(conf.Client.Endpoint),
		otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
			Enabled:         conf.Client.Retry.Enabled,
			InitialInterval: time.Duration(conf.Client.Retry.InitialInterval) * time.Second,
			MaxInterval:     time.Duration(conf.Client.Retry.MaxInterval) * time.Second,
			MaxElapsedTime:  time.Duration(conf.Client.Retry.MaxElapsedTime) * time.Second,
		}),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithCompressor("gzip"),
	)

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(version),
		),
		resource.WithFromEnv(),
		resource.WithProcess(),
	)
	if err != nil {
		logger.Error("failed to create tracer resource: ", err.Error())
	}
	traceExp, err := otlptrace.New(ctx, traceClient)
	if err != nil {
		logger.Error("failed to create tracer exporter: ", err.Error())
	}
	bsp := sdktrace.NewBatchSpanProcessor(traceExp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(conf.SampleRate))),
	)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tracerProvider)
	return func() {
		cxt, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := traceExp.Shutdown(cxt); err != nil {
			otel.Handle(err)
		}
	}
}
