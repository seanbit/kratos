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

// PrepareMiddleWare returns the default middleware stack.
func PrepareMiddleWare() []middleware.Middleware {
	return NewMiddlewareBuilder().Build()
}

// MiddlewareBuilder provides a configurable middleware stack builder.
type MiddlewareBuilder struct {
	enableMetrics    bool
	enableTracing    bool
	enableTraceId    bool
	enableLogging    bool
	enableValidation bool
	enableRateLimit  bool
	enableRecovery   bool
	enableSentry     bool
	extra            []middleware.Middleware
}

// NewMiddlewareBuilder creates a builder with all default middleware enabled.
func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{
		enableMetrics:    true,
		enableTracing:    true,
		enableTraceId:    true,
		enableLogging:    true,
		enableValidation: true,
		enableRateLimit:  true,
		enableRecovery:   true,
		enableSentry:     true,
	}
}

func (b *MiddlewareBuilder) DisableMetrics() *MiddlewareBuilder    { b.enableMetrics = false; return b }
func (b *MiddlewareBuilder) DisableTracing() *MiddlewareBuilder    { b.enableTracing = false; return b }
func (b *MiddlewareBuilder) DisableTraceId() *MiddlewareBuilder    { b.enableTraceId = false; return b }
func (b *MiddlewareBuilder) DisableLogging() *MiddlewareBuilder    { b.enableLogging = false; return b }
func (b *MiddlewareBuilder) DisableValidation() *MiddlewareBuilder { b.enableValidation = false; return b }
func (b *MiddlewareBuilder) DisableRateLimit() *MiddlewareBuilder  { b.enableRateLimit = false; return b }
func (b *MiddlewareBuilder) DisableRecovery() *MiddlewareBuilder   { b.enableRecovery = false; return b }
func (b *MiddlewareBuilder) DisableSentry() *MiddlewareBuilder     { b.enableSentry = false; return b }

// Use appends custom middleware to the stack.
func (b *MiddlewareBuilder) Use(mws ...middleware.Middleware) *MiddlewareBuilder {
	b.extra = append(b.extra, mws...)
	return b
}

// Build constructs the final middleware stack.
func (b *MiddlewareBuilder) Build() []middleware.Middleware {
	var mws []middleware.Middleware
	if b.enableMetrics {
		mws = append(mws, metrics.Server(
			metrics.WithSeconds(_metricSeconds),
			metrics.WithRequests(_metricRequests),
		))
	}
	if b.enableTracing {
		mws = append(mws, tracing.Server())
	}
	if b.enableTraceId {
		mws = append(mws, WriteResponseHeaderTraceId())
	}
	if b.enableLogging {
		mws = append(mws, ServerLogging())
	}
	if b.enableValidation {
		mws = append(mws, validate.ProtoValidate())
	}
	if b.enableRateLimit {
		mws = append(mws, ratelimit.Server())
	}
	if b.enableRecovery {
		mws = append(mws, recovery.Recovery())
	}
	if b.enableSentry {
		// Sentry must be after Recovery, because the exiting order is reversed
		mws = append(mws, sentrykratos.Server())
	}
	mws = append(mws, b.extra...)
	return mws
}
