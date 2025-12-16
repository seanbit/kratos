package server

import (
	"github.com/seanbit/kratos/template/api/web"
	"github.com/seanbit/kratos/template/internal/conf"
	"github.com/seanbit/kratos/template/internal/service"
	"github.com/seanbit/kratos/webkit"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, probe *service.ProbeService, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	middlewares := webkit.PrepareMiddleWare()
	opts = append(opts, grpc.Middleware(middlewares...))
	srv := grpc.NewServer(opts...)
	web.RegisterProbeServer(srv, probe)
	return srv
}
