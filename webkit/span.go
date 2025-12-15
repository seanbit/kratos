package webkit

import (
	"context"
	"time"
)

const (
	CtxApiKey = "ctx_api_name"
)

func CtxSetPath(ctx context.Context, api string) context.Context {
	return context.WithValue(ctx, CtxApiKey, api)
}

func CtxGetPath(ctx context.Context) string {
	if v, ok := ctx.Value(CtxApiKey).(string); ok {
		return v
	}
	return ""
}

func Start() int64 {
	return time.Now().UnixNano()
}

type Span struct {
	ServiceName string
}

func NewSpan(serviceName string) *Span {
	return &Span{serviceName}
}

func (s *Span) EmitCost(ctx context.Context, start int64, proc string) {
	RecordMethodDurationMetricWithCtx(ctx, s.ServiceName, CtxGetPath(ctx), proc, float64(time.Now().UnixNano()-start))
	//MethodDurationMetric.WithLabelValues(s.ServiceName, CtxGetPath(ctx), proc).Observe(float64(time.Now().UnixNano() - start))
}
