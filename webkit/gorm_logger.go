package webkit

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

const LOGGER_PREFIX = "POSTGRES"

var _ gormlogger.Interface = &GormLogger{}

type GormLogger struct {
	klogger       log.Logger
	level         gormlogger.LogLevel
	slowThreshold time.Duration
}

func NewGormLogger(l log.Logger) *GormLogger {
	return &GormLogger{
		klogger:       l,
		level:         gormlogger.Warn,
		slowThreshold: 200 * time.Millisecond,
	}
}

func NewGormLoggerWithArgs(l log.Logger, level gormlogger.LogLevel, slowThreshold time.Duration) *GormLogger {
	// 本地推荐 使用info
	// 4 info : error + warn +  (error sql + 慢sql ) + info
	// 3 warn:  error + warn +  (error sql + 慢sql )
	// 2 error: error + error sql
	return &GormLogger{
		klogger:       l,
		level:         level,
		slowThreshold: slowThreshold,
	}
}

func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newlogger := *l
	newlogger.level = level
	return &newlogger
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Info {
		log.Context(ctx).Infof(msg, data...)
	}
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Warn {
		log.Context(ctx).Warnf(msg, data...)
	}
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Error {
		log.Context(ctx).Errorf(msg, data...)
	}
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}

	config := gormlogger.Config{
		SlowThreshold:             l.slowThreshold,
		IgnoreRecordNotFoundError: true,
	}

	elapsed := time.Since(begin)
	fileNumList := strings.Split(utils.FileWithLineNum(), "/")
	fileNum := fileNumList[len(fileNumList)-1]
	errString := fmt.Sprintf("%+v", err)
	switch {
	case err != nil && l.level >= gormlogger.Error && (!errors.Is(err, gormlogger.ErrRecordNotFound) || !config.IgnoreRecordNotFoundError):
		sql, rows := fc()
		sql = FirstN(sql, 500)
		log.Context(ctx).Errorw(
			"model", LOGGER_PREFIX, "err", errString, "caller", fileNum, "sql_duration_ms", time.Duration(elapsed.Nanoseconds())/time.Millisecond, "sql", sql, "affected_rows", rows,
		)

	case elapsed > config.SlowThreshold && config.SlowThreshold != 0 && l.level >= gormlogger.Warn:
		sql, rows := fc()
		sql = FirstN(sql, 500)
		slowLog := fmt.Sprintf("SLOW SQL >= %v", config.SlowThreshold)
		log.Context(ctx).Warnw(
			"model", LOGGER_PREFIX, "err", errString, "caller", fileNum, "fileNum", fileNum, "slowLog", slowLog, "sql_duration_ms", time.Duration(elapsed.Nanoseconds())/time.Millisecond, "sql", sql, "affected_rows", rows,
		)
	case l.level == gormlogger.Info:
		sql, rows := fc()
		sql = FirstN(sql, 500)
		log.Context(ctx).Infow(
			"model", LOGGER_PREFIX, "err", errString, "caller", fileNum, "sql_duration_ms", time.Duration(elapsed.Nanoseconds())/time.Millisecond, "sql", sql, "affected_rows", rows,
		)
	}

}
