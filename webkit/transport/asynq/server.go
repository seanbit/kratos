// transport/asynq/server.go
package asynq

import (
	"context"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/hibiken/asynq"
)

var _ transport.Server = (*Server)(nil)

// Server Asynq 服务器
type Server struct {
	*asynq.Server
	client *asynq.Client

	config *Config
	logger *log.Helper

	once sync.Once
}

// NewServer 创建 Asynq 服务器
func NewServer(opts ...ServerOption) *Server {
	config := &Config{
		RedisURI:    "redis://127.0.0.1:6379/9",
		Concurrency: 10,
		Queues: map[string]int{
			"default": 1,
		},
		Logger: log.GetLogger(),
	}

	for _, opt := range opts {
		opt(config)
	}

	return &Server{
		config: config,
		logger: log.NewHelper(config.Logger),
	}
}

// Start 启动 Asynq 服务器
func (s *Server) Start(ctx context.Context) error {
	// 初始化 Redis 连接
	redisConnOpts, err := asynq.ParseRedisURI(s.config.RedisURI)
	if err != nil {
		return err
	}
	s.once.Do(func() {
		// 创建 Asynq 服务器
		s.Server = asynq.NewServer(
			redisConnOpts,
			asynq.Config{
				Concurrency: s.config.Concurrency,
				Queues:      s.config.Queues,
				Logger:      NewAsynqLogger(s.logger),
			},
		)

		// 创建 Asynq 客户端
		s.client = asynq.NewClient(redisConnOpts)
	})

	// 在 goroutine 中启动服务器
	go func() {
		s.logger.Infof("Asynq server starting with: %s", s.config.RedisURI)
		if err := s.Server.Start(s.config.Handler); err != nil {
			s.logger.Errorf("Asynq server run error: %v", err)
		}
	}()
	return nil
}

// Stop 停止 Asynq 服务器
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Asynq server stopping")

	if s.Server != nil {
		s.Server.Shutdown()
	}

	if s.client != nil {
		s.client.Close()
	}

	s.logger.Info("Asynq server stopped")
	return nil
}

// Client 获取 Asynq 客户端
func (s *Server) Client() *asynq.Client {
	return s.client
}
