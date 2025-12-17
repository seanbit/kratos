package server

import (
	"github.com/google/wire"
	"github.com/seanbit/kratos/template/internal/server/middlewares"
	"github.com/seanbit/kratos/webkit/transport/crontab"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(
	middlewares.NewUserAuth,
	middlewares.NewHttpBuilder,
	NewGRPCServer,
	NewHTTPServer,
	NewAsynqServer,
	NewAsynqClient,
	crontab.NewServer,
)
