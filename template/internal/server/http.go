package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kratos/kratos/v2/transport"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/seanbit/kratos/template/api/web"
	"github.com/seanbit/kratos/template/internal/biz"
	"github.com/seanbit/kratos/template/internal/conf"
	"github.com/seanbit/kratos/template/internal/server/middlewares"
	"github.com/seanbit/kratos/template/internal/service"
	"github.com/seanbit/kratos/webkit"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, logger log.Logger, middlewaresBuilder *middlewares.HttpBuilder,
	probe *service.ProbeService, alarm biz.IAlarmRepo, auth *service.AuthService,
) *khttp.Server {
	var opts = []khttp.ServerOption{
		khttp.Filter(handlers.CORS(
			handlers.AllowedOriginValidator(func(origin string) bool {
				// 这里可以添加更多的逻辑来决定是否允许这个origin
				return true
			}),
			handlers.AllowedHeaders([]string{
				"Access-Control-Allow-Credentials",
				"authorization",
				"Authorization",
				"Content-Length",
				"Content-Type",
				"DNT",
				"Keep-Alive",
				"Origin",
				"token",
				"tmp_user_id",
				"User-Agent",
				"x-admin-id",
				"X-App-Id",
				"X-Requested-With",
				"Debug-Summary",
			}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowCredentials(),
			handlers.MaxAge(600),
		)),
		khttp.ResponseEncoder(webkit.ResponseEncoder),
		khttp.ErrorEncoder(webkit.ErrorEncoder(web.ErrorReason_value)),
	}
	if c.Http.Network != "" {
		opts = append(opts, khttp.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, khttp.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, khttp.Timeout(c.Http.Timeout.AsDuration()))
	}

	middlewareFns := webkit.PrepareMiddleWare()
	middlewareFns = append(middlewareFns, InjectContextMiddleware())
	recoverFunc := func(ctx context.Context, req, err interface{}) error {
		alarm.SendBizMessage(ctx, "panic error", "panic error")
		log.Context(ctx).Errorf("panic error")
		return recovery.ErrUnknownRequest
	}
	middlewareFns = append(middlewareFns, recovery.Recovery(recovery.WithHandler(recoverFunc)))
	middlewareFns = append(middlewareFns,
		middlewaresBuilder.Build()...,
	)
	opts = append(opts, khttp.Middleware(middlewareFns...))

	srv := khttp.NewServer(opts...)
	web.RegisterProbeHTTPServer(srv, probe)
	web.RegisterAuthHTTPServer(srv, auth)
	srv.Handle("/metrics", promhttp.Handler())

	// 注册健康检查端点（返回详细状态，供监控系统使用）
	srv.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		status := probe.GetHealthStatus(r.Context())
		w.Header().Set("Content-Type", "application/json")
		if status.Status != "healthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_ = json.NewEncoder(w).Encode(status)
	})

	return srv
}

func InjectContextMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				if ht, ok := tr.(khttp.Transporter); ok {
					path := ht.Request().URL.Path
					ctx = webkit.CtxSetPath(ctx, path)
				}
			}
			reply, err = handler(ctx, req)
			return
		}
	}
}
