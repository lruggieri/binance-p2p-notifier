package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Default *Logger

type Logger struct {
	*zap.Logger
}

func (l *Logger) WithField(field string, value interface{}) *Logger {
	return &Logger{
		Logger: l.Logger.With(zap.Field{
			Key:       field,
			Type:      zapcore.ReflectType,
			Interface: value,
		}),
	}
}

func (l *Logger) WithError(err error) *Logger {
	return l.WithField("error", err)
}

func InitDefault() {
	zl, _ := zap.NewProduction()
	Default = &Logger{
		Logger: zl,
	}
}
