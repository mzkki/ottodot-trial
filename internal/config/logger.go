package config

import (
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(config *viper.Viper) (*zap.Logger, error) {
	zapCfg := zap.NewProductionConfig()

	zapCfg.Encoding = "json"
	zapCfg.OutputPaths = []string{"stdout"}
	zapCfg.ErrorOutputPaths = []string{"stderr"}

	level := strings.ToLower(config.GetString("LOG_LEVEL"))
	switch level {
	case "debug":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "warn":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	zapCfg.EncoderConfig.TimeKey = "timestamp"
	zapCfg.EncoderConfig.LevelKey = "level"
	zapCfg.EncoderConfig.MessageKey = "message"
	zapCfg.EncoderConfig.CallerKey = "caller"
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	base, err := zapCfg.Build(zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		return nil, err
	}

	return base.With(
		zap.String("service", config.GetString("SERVICE_NAME")),
		zap.String("environment", config.GetString("APP_ENV")),
	), nil
}
