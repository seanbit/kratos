// transport/asynq/config.go
package asynq

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

// Config Asynq 配置
type Config struct {
	RedisURI    string         `json:"redis_uri"`
	Concurrency int            `json:"concurrency"`
	Queues      map[string]int `json:"queues"`
	Logger      log.Logger     `json:"-"`
	Handler     asynq.Handler  `json:"-"`
}

// ServerOption Asynq 服务器选项
type ServerOption func(*Config)

// WithRedisURI 设置 Redis 配置
func WithRedisURI(uri string) ServerOption {
	return func(c *Config) {
		c.RedisURI = uri
	}
}

// WithConcurrency 设置并发数
func WithConcurrency(concurrency int) ServerOption {
	return func(c *Config) {
		c.Concurrency = concurrency
	}
}

// WithQueues 设置队列配置
func WithQueues(queues map[string]int) ServerOption {
	return func(c *Config) {
		c.Queues = queues
	}
}

// WithLogger 设置队列配置
func WithLogger(logger log.Logger) ServerOption {
	return func(c *Config) {
		c.Logger = logger
	}
}

func WithHandler(h asynq.Handler) ServerOption {
	return func(c *Config) {
		c.Handler = h
	}
}
