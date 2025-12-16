//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/seanbit/kratos/template/internal/biz"
	"github.com/seanbit/kratos/template/internal/conf"
	"github.com/seanbit/kratos/template/internal/data"
	"github.com/seanbit/kratos/template/internal/infra"
	"github.com/seanbit/kratos/template/internal/server"
	"github.com/seanbit/kratos/template/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.S3, *conf.GeoIp, *conf.Alarm, *conf.Auth, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(infra.ProviderSet, server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
