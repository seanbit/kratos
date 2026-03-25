//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"evm-scan/internal/conf"
	"evm-scan/internal/crontab"
)

// wireApp init kratos application.
func wireApp(*conf.CronJob, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(crontab.ProviderSet, newApp))
}
