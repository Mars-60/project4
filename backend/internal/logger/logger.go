package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func Init(env string) error {

	return InitWithLevel(env, "info")
}

func InitWithLevel(env string, level string) error {
	atomicLevel := zap.NewAtomicLevelAt(parseLevel(level))
	var cfg zap.Config
	var err error

	if env == "development" {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	cfg.Level = atomicLevel
	logger, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = logger

	return nil
}

func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
