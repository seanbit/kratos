package webkit

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"go.opentelemetry.io/otel/trace"
)

func WriteResponseHeaderTraceId() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				if span := trace.SpanContextFromContext(ctx); span.HasTraceID() {
					tr.ReplyHeader().Set("X-Trace-Id", span.TraceID().String())
				}
			}

			return handler(ctx, req)
		}
	}
}
