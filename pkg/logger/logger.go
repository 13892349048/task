package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger
	once         sync.Once
)

// Init initializes a global zap logger with the given environment (dev/prod).
func Init(environment string) (*zap.Logger, error) {
	var cfg zap.Config
	if environment == "prod" || environment == "production" {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "ts"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}
	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	once.Do(func() {
		globalLogger = l
	})
	return l, nil
}

// L returns the global logger. Should call Init first.
func L() *zap.Logger {
	if globalLogger == nil {
		l, _ := zap.NewDevelopment()
		globalLogger = l
	}
	return globalLogger
}
