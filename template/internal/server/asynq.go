package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	asynq2 "github.com/hibiken/asynq"
	"github.com/seanbit/kratos/template/api/event"
	"github.com/seanbit/kratos/template/internal/conf"
	"github.com/seanbit/kratos/webkit/transport/asynq"
)

func NewAsynqClient(config *conf.Server) (*asynq2.Client, error) {
	redisConnOpts, err := asynq2.ParseRedisURI(config.Asynq.RedisUri)
	if err != nil {
		return nil, err
	}
	return asynq2.NewClient(redisConnOpts), nil
}

func NewAsynqServer(config *conf.Server, logger log.Logger, handler event.EventHandlerServer) *asynq.Server {
	opts := []asynq.ServerOption{
		asynq.WithRedisURI(config.Asynq.RedisUri),
		asynq.WithLogger(logger),
		asynq.WithHandler(newAsynqProcesser(handler)),
	}
	if config.Asynq.Concurrency > 0 {
		opts = append(opts, asynq.WithConcurrency(int(config.Asynq.Concurrency)))
	}
	if len(config.Asynq.Queues) > 0 {
		queues := make(map[string]int, len(config.Asynq.Queues))
		for name, queue := range config.Asynq.Queues {
			queues[name] = int(queue)
		}
		opts = append(opts, asynq.WithQueues(queues))
	}
	return asynq.NewServer(opts...)
}

type AsynqProcesser struct {
	handler event.EventHandlerServer
}

func newAsynqProcesser(handler event.EventHandlerServer) *AsynqProcesser {
	return &AsynqProcesser{handler: handler}
}

func (p *AsynqProcesser) ProcessTask(ctx context.Context, task *asynq2.Task) error {
	_, err := p.handler.HandleEvent(ctx, &event.Event{
		Id:      task.ResultWriter().TaskID(),
		Name:    task.Type(),
		Payload: task.Payload(),
	})
	return err
}
