package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type Probe struct {
}

func NewProbe() *Probe { return &Probe{} }

// Ready 检查进程是否完成所有的初始化
func (biz *Probe) Ready(ctx context.Context, in map[string]interface{}) error {
	log.Context(ctx).Debug("receive Readiness probe")

	// TODO 进行必要的检查

	return nil
}
