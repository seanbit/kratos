//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/seanbit/kratos/template/internal/biz"
	"github.com/seanbit/kratos/template/internal/conf"
	"github.com/seanbit/kratos/template/internal/crontab"
	"github.com/seanbit/kratos/template/internal/data"
	"github.com/seanbit/kratos/template/internal/infra"
	"github.com/seanbit/kratos/template/internal/server"
	"github.com/seanbit/kratos/template/internal/service"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.S3, *conf.GeoIp, *conf.Alarm, *conf.Auth, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(infra.ProviderSet, data.ProviderSet, biz.ProviderSet, crontab.ProviderSet, service.ProviderSet, server.ProviderSet, newApp))
}
