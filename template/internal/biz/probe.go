package biz

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

// HealthStatus 健康检查状态（biz层定义）
type HealthStatus struct {
	Status     string                      `json:"status"` // "healthy" or "unhealthy"
	Components map[string]*ComponentHealth `json:"components"`
}

// ComponentHealth 组件健康状态
type ComponentHealth struct {
	Status   string                 `json:"status"`              // "healthy" or "unhealthy"
	Latency  string                 `json:"latency,omitempty"`   // 响应时间
	Error    string                 `json:"error,omitempty"`     // 错误信息
	StatsExt map[string]interface{} `json:"stats_ext,omitempty"` // 状态扩展
}

// IHealthRepo 健康检查仓库接口（由data层实现）
type IHealthRepo interface {
	CheckHealth(ctx context.Context) *HealthStatus
}

type Probe struct {
	healthRepo IHealthRepo
}

func NewProbe(healthRepo IHealthRepo) *Probe {
	return &Probe{healthRepo: healthRepo}
}

// Ready 检查进程是否完成所有的初始化
func (p *Probe) Ready(ctx context.Context, in map[string]interface{}) error {
	log.Context(ctx).Debug("receive Readiness probe")

	status := p.healthRepo.CheckHealth(ctx)
	if status.Status != "healthy" {
		var errMsg string
		for name, comp := range status.Components {
			if comp.Status == "unhealthy" {
				errMsg += fmt.Sprintf("%s: %s; ", name, comp.Error)
			}
		}
		return fmt.Errorf("readiness check failed: %s", errMsg)
	}

	return nil
}

// CheckHealth 返回详细的健康状态
func (p *Probe) CheckHealth(ctx context.Context) *HealthStatus {
	return p.healthRepo.CheckHealth(ctx)
}
