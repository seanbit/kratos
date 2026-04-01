//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"github.com/seanbit/kratos/template/cmd/job/jobs"
	"github.com/seanbit/kratos/template/internal/conf"
	"github.com/seanbit/kratos/template/internal/infra"
)

// initApp 初始化应用依赖
func initApp(*conf.Data, log.Logger) (*jobs.App, func(), error) {
	panic(wire.Build(
		// 基础设施层
		infra.ProviderSet,

		// Job App
		newJobApp,
	))
}

// newJobApp 创建 Job App
func newJobApp(logger log.Logger) *jobs.App {
	return &jobs.App{
		Logger: logger,
	}
}
