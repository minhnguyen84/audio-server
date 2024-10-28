package utils

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogWrapperObj struct {
	logger *zap.Logger
}

var Logger = LogWrapperObj{
	logger: initLogger(),
}

func initLogger() *zap.Logger {
	callerSkip := zap.AddCallerSkip(1)

	if gin.IsDebugging() {
		config := zap.NewDevelopmentConfig()
		config.OutputPaths = []string{"stdout"}
		logger, _ := config.Build()
		return logger.WithOptions(callerSkip)
	} else {
		config := zap.NewProductionConfig()
		config.Level.SetLevel(zap.InfoLevel)
		config.EncoderConfig.LevelKey = "severity"
		config.EncoderConfig.EncodeLevel = LevelGcpEncoding
		config.EncoderConfig.TimeKey = "time"
		config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		config.OutputPaths = []string{"stdout"}
		logger, _ := config.Build()
		return logger.With(zap.Namespace("app")).WithOptions(callerSkip).With(zap.String("version", AppVersion))
	}
}

func LevelGcpEncoding(level zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {
	encoder.AppendString(convertLevel(level))
}

func convertLevel(level zapcore.Level) string {
	switch level {
	case zapcore.DebugLevel:
		return "DEBUG"
	case zapcore.InfoLevel:
		return "INFO"
	case zapcore.WarnLevel:
		return "WARNING"
	case zapcore.ErrorLevel:
		return "ERROR"
	case zapcore.DPanicLevel:
		return "CRITICAL"
	case zapcore.PanicLevel:
		return "CRITICAL"
	case zapcore.FatalLevel:
		return "EMERGENCY"
	default:
		return "DEFAULT"
	}
}

func (logWrapper LogWrapperObj) Info(message string, fields ...zap.Field) {
	logWrapper.logger.Info(message, fields...)
}

func (logWrapper LogWrapperObj) Warn(message string, fields ...zap.Field) {
	logWrapper.logger.Warn(message, fields...)
}

func (logWrapper LogWrapperObj) Error(message string, fields ...zap.Field) {
	logWrapper.logger.Error(message, fields...)
}

func (logWrapper LogWrapperObj) Fatal(message string, fields ...zap.Field) {
	logWrapper.logger.Fatal(message, fields...)
}
