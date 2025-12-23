package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config defines logger settings.
type Config struct {
	Level       string
	Development bool
}

// New builds a zap logger based on config.
func New(cfg Config) (*zap.Logger, error) {
	zapCfg := zap.NewProductionConfig()
	if cfg.Development {
		zapCfg = zap.NewDevelopmentConfig()
	}

	level := zapcore.InfoLevel
	if cfg.Level != "" {
		if err := level.Set(strings.ToLower(cfg.Level)); err != nil {
			return nil, err
		}
	}
	zapCfg.Level = zap.NewAtomicLevelAt(level)
	return zapCfg.Build()
}
