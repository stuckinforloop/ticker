package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(env string) (*zap.Logger, error) {
	var cfg zap.Config

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
