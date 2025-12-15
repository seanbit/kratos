package webkit

import (
	"github.com/go-kratos/kratos/contrib/middleware/validate/v2"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	sentrykratos "github.com/go-kratos/sentry"
)

func PrepareMiddleWare() []middleware.Middleware {
	return []middleware.Middleware{
		metrics.Server(
			metrics.WithSeconds(_metricSeconds),
			metrics.WithRequests(_metricRequests),
		),
		tracing.Server(),
		WriteResponseHeaderTraceId(),
		ServerLogging(),
		validate.ProtoValidate(),
		ratelimit.Server(),
		recovery.Recovery(),
		sentrykratos.Server(), // must after Recovery middleware, because of the exiting order will be reversed
	}
}
