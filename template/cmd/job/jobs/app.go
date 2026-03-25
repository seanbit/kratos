package jobs

import "evm-scan/internal/biz"

// App 依赖容器（由 Wire 生成）
type App struct {
	EventRepo       biz.IEventRepo
	EventDispatcher biz.IEventDispatcher
}
