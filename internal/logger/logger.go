package logger

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New() (*zap.Logger, error) {
	var cfg zap.Config

	env := viper.GetString("environment")

	switch env {
	case "dev":
		cfg = zap.NewDevelopmentConfig()
	case "prod":
		cfg = zap.NewProductionConfig()
	}

	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.MessageKey = "message"
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	return cfg.Build()
}
