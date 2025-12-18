package main

import (
	microzap "github.com/go-micro/plugins/v4/logger/zap"
	"github.com/zhanshen02154/product/internal/bootstrap"
	configstruct "github.com/zhanshen02154/product/internal/config"
	"go-micro.dev/v4/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	_ "net/http/pprof"
	"os"
)

func main() {
	loggerMetadataMap := make(map[string]interface{})
	zapLogger := zap.New(zapcore.NewCore(getEncoder(), zapcore.AddSync(os.Stdout), zap.InfoLevel),
		zap.WithCaller(true),
		zap.AddCallerSkip(1),
	)
	defer zapLogger.Sync()
	microLogger, err := microzap.NewLogger(microzap.WithLogger(zapLogger))
	if err != nil {
		log.Println(err)
		return
	}
	logger.DefaultLogger = microLogger

	// 从consul获取配置
	conf, err := configstruct.GetConfig()
	if err != nil {
		logger.Error("service load config fail: ", err)
		return
	}

	var confInfo configstruct.SysConfig
	if err = conf.Get("product").Scan(&confInfo); err != nil {
		logger.Error(err)
		return
	}
	componentLogger := zapLogger.With(
		zap.String("service", confInfo.Service.Name),
		zap.String("version", confInfo.Service.Version),
	)
	loggerMetadataMap["service"] = confInfo.Service.Name
	loggerMetadataMap["version"] = confInfo.Service.Version
	loggerMetadataMap["type"] = "core"
	logger.DefaultLogger = logger.DefaultLogger.Fields(loggerMetadataMap)

	if err := bootstrap.RunService(&confInfo, componentLogger); err != nil {
		logger.Error("failed to start service: ", err)
	}
}

// 获取日志编码器
func getEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(
		zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			TimeKey:        "timestamp",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.MillisDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		})
}
