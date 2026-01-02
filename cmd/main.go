package main

import (
	"context"
	"fmt"
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
		zapLogger.Error(fmt.Sprintf("failed to load config: %s", err.Error()))
		return
	}

	var confInfo configstruct.SysConfig
	if err := consulSource.Get("product").Scan(&confInfo); err != nil {
		zapLogger.Error(fmt.Sprintf("failed convert config to struct: %s", err.Error()))
		return
	}
	// 检查配置
	if err := confInfo.CheckConfig(); err != nil {
		zapLogger.Error(fmt.Sprintf("failed to check config: %s", err.Error()))
		return
	}
	serverLogLevel := infrastructure.FindZapLogLevel(confInfo.Service.LogLevel)
	componentLogger := zapLogger.WithOptions(
		zap.IncreaseLevel(serverLogLevel),
		zap.Fields(zap.String("service", confInfo.Service.Name)),
		zap.Fields(zap.String("version", confInfo.Service.Version)),
	)
	loggerMetadataMap := make(map[string]interface{})
	loggerMetadataMap["type"] = "core"
	microLogger, err := microzap.NewLogger(
		microzap.WithLogger(componentLogger),
		logger.WithFields(loggerMetadataMap),
	)
	if err != nil {
		zapLogger.Error(fmt.Sprintf("failed to load go micro logger: %s", err.Error()))
		return
	}
	logger.DefaultLogger = microLogger

	// 链路追踪
	traceShutdown := initTracer(confInfo.Service.Name, confInfo.Service.Version, confInfo.Tracer)
	defer traceShutdown()

	if err := bootstrap.RunService(&confInfo, componentLogger); err != nil {
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
