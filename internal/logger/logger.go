package logger

import (
	"strings"

	"go.uber.org/zap"
)

type Logger struct {
	*zap.Logger
}

func NewLogger(level, format string) *Logger {
	var config zap.Config

	switch strings.ToLower(format) {
	case "json":
		config = zap.NewProductionConfig()
	case "console":
		config = zap.NewDevelopmentConfig()
	default:
		config = zap.NewProductionConfig()
	}

	// Set log level
	switch strings.ToLower(level) {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return &Logger{Logger: logger}
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}

func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{Logger: l.Logger.With(fields...)}
}

func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
