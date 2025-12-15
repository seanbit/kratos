// transport/asynq/logger.go
package asynq

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

// AsynqLogger Asynq 日志适配器
type AsynqLogger struct {
	logger *log.Helper
}

// NewAsynqLogger 创建 Asynq 日志适配器
func NewAsynqLogger(logger *log.Helper) asynq.Logger {
	return &AsynqLogger{logger: logger}
}

func (l *AsynqLogger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

func (l *AsynqLogger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *AsynqLogger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l *AsynqLogger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l *AsynqLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}
