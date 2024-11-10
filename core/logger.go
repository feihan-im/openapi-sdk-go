package fhcore

import (
	"context"
	"log"
	"os"
)

type LoggerLevel int8

const (
	LoggerLevelDebug LoggerLevel = iota
	LoggerLevelInfo
	LoggerLevelWarn
	LoggerLevelError
)

type Logger interface {
	Debugf(ctx context.Context, format string, args ...interface{})
	Infof(ctx context.Context, format string, args ...interface{})
	Warnf(ctx context.Context, format string, args ...interface{})
	Errorf(ctx context.Context, format string, args ...interface{})
}

func NewDefaultLogger(level LoggerLevel) Logger {
	return &defaultLogger{
		level:  level,
		logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
	}
}

type defaultLogger struct {
	level  LoggerLevel
	logger *log.Logger
}

func (l defaultLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	if l.level <= LoggerLevelDebug {
		l.logger.Printf("[feihan debug] "+format, args...)
	}
}

func (l defaultLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	if l.level <= LoggerLevelInfo {
		l.logger.Printf("[feihan info] "+format, args...)
	}
}

func (l defaultLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	if l.level <= LoggerLevelWarn {
		l.logger.Printf("[feihan warn] "+format, args...)
	}
}

func (l defaultLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	if l.level <= LoggerLevelError {
		l.logger.Printf("[feihan error] "+format, args...)
	}
}
