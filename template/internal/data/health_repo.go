package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/seanbit/kratos/template/internal/biz"
	"github.com/seanbit/kratos/template/internal/infra"
)

type HealthRepo struct {
	dbProvider  infra.PostgresProvider
	rdbProvider infra.RedisProvider
	log         *log.Helper
}

func NewHealthRepo(dbProvider infra.PostgresProvider, rdbProvider infra.RedisProvider, logger log.Logger) biz.IHealthRepo {
	return &HealthRepo{
		dbProvider:  dbProvider,
		rdbProvider: rdbProvider,
		log:         log.NewHelper(logger),
	}
}

// CheckHealth 检查所有组件的健康状态
func (r *HealthRepo) CheckHealth(ctx context.Context) *biz.HealthStatus {
	status := &biz.HealthStatus{
		Status:     "healthy",
		Components: make(map[string]*biz.ComponentHealth),
	}

	// 检查 PostgreSQL
	status.Components["postgres"] = r.checkPostgres(ctx)
	if status.Components["postgres"].Status == "unhealthy" {
		status.Status = "unhealthy"
	}

	// 检查 Redis
	status.Components["redis"] = r.checkRedis(ctx)
	if status.Components["redis"].Status == "unhealthy" {
		status.Status = "unhealthy"
	}

	return status
}

// checkPostgres 检查 PostgreSQL 连接
func (r *HealthRepo) checkPostgres(ctx context.Context) *biz.ComponentHealth {
	health := &biz.ComponentHealth{Status: "healthy"}

	start := time.Now()
	db, err := r.dbProvider.GetDB().DB()
	if err != nil {
		health.Status = "unhealthy"
		health.Error = err.Error()
		return health
	}

	if err := db.PingContext(ctx); err != nil {
		health.Status = "unhealthy"
		health.Error = err.Error()
		return health
	}

	health.Latency = time.Since(start).String()
	return health
}

// checkRedis 检查 Redis 连接
func (r *HealthRepo) checkRedis(ctx context.Context) *biz.ComponentHealth {
	health := &biz.ComponentHealth{Status: "healthy"}

	start := time.Now()
	if err := r.rdbProvider.GetRedis().Ping(ctx).Err(); err != nil {
		health.Status = "unhealthy"
		health.Error = err.Error()
		return health
	}

	health.Latency = time.Since(start).String()
	return health
}
