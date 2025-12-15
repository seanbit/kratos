package webkit

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
)

// Redacter defines how to log an object
type Redacter interface {
	Redact() string
}

// Server is an server logging middleware.
func ServerLogging() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			startTime := time.Now()
			reply, err = handler(ctx, req)
			// 这里对err 统一处理为*errors.Error
			var se *errors.Error
			var ok bool
			if se, ok = err.(*errors.Error); !ok {
				// 兼容errors.New("...")
				if err != nil {
					se = errors.Newf(5000, err.Error(), err.Error())
				} else {
					se = errors.Newf(0, "", "")
				}
			}
			args := FirstN(extractArgs(req), 500)
			log.Context(ctx).Infow(
				"args", args,
				"code", se.Code,
				"msg", se.Message,
				"latency", time.Since(startTime).Seconds(),
			)
			return
		}
	}
}

// Client is a client logging middleware.
func ClientLogging(logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			var (
				code int32
				msg  string
			)
			startTime := time.Now()
			reply, err = handler(ctx, req)
			if se := errors.FromError(err); se != nil {
				code = se.Code
				msg = se.Message
			}
			_, stack := extractError(err)
			args := FirstN(extractArgs(req), 500)
			log.Context(ctx).Infow(
				"args", args,
				"code", code,
				"msg", msg,
				"stack", stack,
				"latency", time.Since(startTime).Seconds(),
			)
			return
		}
	}
}

// extractArgs returns the string of the req
func extractArgs(req interface{}) string {
	if redacter, ok := req.(Redacter); ok {
		return redacter.Redact()
	}
	if stringer, ok := req.(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprintf("%+v", req)
}

// extractError returns the string of the error
func extractError(err error) (log.Level, string) {
	if err != nil {
		return log.LevelError, fmt.Sprintf("%+v", err)
	}
	return log.LevelInfo, ""
}
