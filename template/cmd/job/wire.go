//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"github.com/seanbit/kratos/template/cmd/job/jobs"
)

// initApp 初始化应用依赖
func initApp(log.Logger) (*jobs.App, func(), error) {
	panic(wire.Build(
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
