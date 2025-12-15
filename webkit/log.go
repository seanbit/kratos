package webkit

import (
	"context"
	"errors"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	kratoszero "github.com/go-kratos/kratos/contrib/log/zerolog/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/rs/zerolog"
)

var logger log.Logger

func GetLogger() log.Logger {
	if logger == nil {
		panic("logger has not been initialized, please call InitLogger first")
	}

	return logger
}

func InitLogger(name, version string, level int) {

	//logger = getZapLogger(level)
	logger = getZeroLogger(level)

	fnSentry := func(level log.Level, keyvals ...interface{}) bool {
		// error和fatal级别的日志上报sentry事件
		if level == log.LevelError || level == log.LevelFatal {
			for i := 0; i < len(keyvals); i++ {
				if keyvals[i] == "stack" || keyvals[i] == "msg" {
					// 忽略特定的error事件
					if strings.HasPrefix(keyvals[i+1].(string), "tips: ") {
						return false
					}
				}
			}

			// pretty message
			evt := sentry.NewEvent()
			for i := 0; i < len(keyvals); i++ {
				switch keyvals[i] {
				// 提高可读性，忽略某些字段
				case "time":
				case "caller":
				case "trace.id":
				case "span.id":
				case "stack":
				case "service.name":
				case "service.version":
				case "msg":
					evt.Message = keyvals[i+1].(string)
				default:
					evt.Extra[keyvals[i].(string)] = keyvals[i+1]
				}
				i++
			}
			if evt.Message == "" {
				return false
			}

			evt.SetException(errors.New(evt.Message), 1)
			evt.Level = sentry.LevelError
			sentry.CaptureEvent(evt)
		}
		return false
	}
	log.SetLogger(
		log.NewFilter(
			log.With(logger,
				"time", log.DefaultTimestamp,
				"service.name", name,
				"trace.id", tracing.TraceID(),
				"span.id", tracing.SpanID(),
				"caller", Caller(5),
				"operation", GetLogOperation(),
				"uid", GetLogUid(),
			),
			log.FilterFunc(fnSentry),
		),
	)

}

func GetLogUid() log.Valuer {
	return func(ctx context.Context) interface{} {
		uid := ""
		defer func() {
			if err := recover(); err != nil {
				return
			}
		}()
		uid, _ = UserIdFromContext(ctx)
		return uid
	}
}

func GetLogOperation() log.Valuer {
	return func(ctx context.Context) interface{} {
		defer func() {
			if err := recover(); err != nil {
				return
			}
		}()
		if info, ok := transport.FromServerContext(ctx); ok {
			return info.Operation()
		}
		return ""
	}
}

func Caller(depth int) log.Valuer {
	return func(context.Context) interface{} {
		pc, file, line, success := runtime.Caller(depth)
		if !success {
			return "file:0:f"
		}
		idx := strings.LastIndexByte(file, '/')
		if idx == -1 {
			return file[idx+1:] + ":" + strconv.Itoa(line)
		}
		idx = strings.LastIndexByte(file[:idx], '/')

		fn := runtime.FuncForPC(pc)
		var fnName string
		if fn == nil {
			fnName = "?()"
		} else {
			funcPathList := strings.Split(runtime.FuncForPC(pc).Name(), "/")
			fnName = funcPathList[len(funcPathList)-1]
		}

		return file[idx+1:] + ":" + strconv.Itoa(line) + ":" + fnName
	}
}

func getZeroLogger(level int) log.Logger {
	var output io.Writer
	if os.Getenv("ENV") == "local" {
		output = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	} else {
		output = os.Stdout
	}

	zlogger := zerolog.New(output)
	logger := kratoszero.NewLogger(&zlogger)

	//enab := zerolog.InfoLevel
	//switch level {
	//case "debug":
	//	enab = zerolog.DebugLevel
	//case "info":
	//	enab = zerolog.InfoLevel
	//case "warn":
	//	enab = zerolog.WarnLevel
	//case "error":
	//	enab = zerolog.ErrorLevel
	//case "panic":
	//	enab = zerolog.PanicLevel
	//case "fatal":
	//	enab = zerolog.FatalLevel
	//}
	zerolog.SetGlobalLevel(zerolog.Level(level))

	return logger
}
