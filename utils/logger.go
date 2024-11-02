package utils

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LogWrapperObj struct {
	logger *zap.Logger
}

var Logger = LogWrapperObj{
	logger: initLogger(),
}

func initLogger() *zap.Logger {
	callerSkip := zap.AddCallerSkip(1)

	appConfig := GetConfig()
	config := zap.NewDevelopmentConfig()
	if appConfig.IsDebugging {
		config.Level.SetLevel(zap.DebugLevel)
	} else {
		gin.SetMode(gin.ReleaseMode)
		config.Level.SetLevel(zap.InfoLevel)
	}
	config.OutputPaths = []string{"stdout"}
	logger, _ := config.Build()
	return logger.WithOptions(callerSkip)
}

func (logWrapper LogWrapperObj) Debug(message string, fields ...zap.Field) {
	logWrapper.logger.Debug(message, fields...)
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
