package logger

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ProvideLogger provides a zap logger
func ProvideLogger() *zap.SugaredLogger {
	logger, _ := zap.NewProduction()
	slogger := logger.Sugar()

	return slogger
}

// Module provided to fx
var Module = fx.Options(
	fx.Provide(ProvideLogger),
)
