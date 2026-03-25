//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"evm-scan/cmd/job/jobs"
	"evm-scan/internal/biz"
	"evm-scan/internal/biz/processor"
	"evm-scan/internal/conf"
	"evm-scan/internal/data"
	"evm-scan/internal/infra"
)

// initApp 初始化应用依赖
func initApp(*conf.Server, *conf.Data, *conf.Blockchain, *conf.Scanner, log.Logger) (*jobs.App, func(), error) {
	panic(wire.Build(
		// 基础设施层
		infra.ProviderSet,

		// 数据层
		data.ProviderSet,

		// 业务层（包含 processor）
		processor.ProviderSet,
		//biz.ProviderSet,

		// Job App
		newJobApp,
	))
}

// newJobApp 创建 Job App
func newJobApp(
	eventRepo biz.IEventRepo,
	eventDispatcher biz.IEventDispatcher,
) *jobs.App {
	return &jobs.App{
		EventRepo:       eventRepo,
		EventDispatcher: eventDispatcher,
	}
}
