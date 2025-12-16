package infra

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/seanbit/kratos/template/internal/conf"
)

const (
	redisDefaultPoolSize    = 60
	redisDefaultMinIdleConn = 20
	redisDefaultMaxRetries  = 3
	redisDefaultIdleTimeout = 5 * time.Minute
)

type RedisProvider interface {
	GetRedis() redis.UniversalClient
	Close()
}

type redisProvider struct {
	config        *conf.Data_Redis
	client        *redis.Client
	clusterClient *redis.ClusterClient
	cleaning      *atomic.Bool
}

func NewRedisProvider(config *conf.Data_Redis) (RedisProvider, func(), error) {
	provider := &redisProvider{config: config, cleaning: &atomic.Bool{}}
	var err error
	if config.IsCluster {
		provider.clusterClient, err = newRedisClusterClient(config)
	} else {
		provider.client, err = newRedisClient(config)
	}
	if err != nil {
		return nil, nil, errors.Wrap(err, "new redis provider")
	}
	return provider, provider.Close, err
}

func (p *redisProvider) GetRedis() redis.UniversalClient {
	if p.config.IsCluster {
		return p.clusterClient
	} else {
		return p.client
	}
}

func (p *redisProvider) Close() {
	if p.cleaning.Load() {
		return
	}
	p.cleaning.Store(true)
	if p.client != nil {
		if err := p.client.Close(); err != nil {
			log.Warnf("redis Close error: %+v", err)
		}
	}
	if p.clusterClient != nil {
		if err := p.clusterClient.Close(); err != nil {
			log.Warnf("redis cluster Close error: %+v", err)
		}
	}
	log.Info("redis closed")
}

func newRedisClient(c *conf.Data_Redis) (*redis.Client, error) {
	poolSize := c.GetPoolSize()
	if poolSize == 0 {
		poolSize = redisDefaultPoolSize
	}
	minIdleConn := c.MinIdleConn
	if minIdleConn <= 0 {
		minIdleConn = redisDefaultMinIdleConn
	}
	maxRetries := c.GetPoolSize()
	if maxRetries <= 0 {
		maxRetries = redisDefaultMaxRetries
	}
	idleTimeout := c.GetIdleTimeout().AsDuration()
	if idleTimeout <= 0 {
		idleTimeout = redisDefaultIdleTimeout
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:            c.Addr,
		Username:        c.Username,
		Password:        c.Password, // no password set
		DB:              int(c.Db),  // use default DB
		MaxRetries:      int(maxRetries),
		ReadTimeout:     c.GetReadTimeout().AsDuration(),
		WriteTimeout:    c.GetWriteTimeout().AsDuration(),
		PoolSize:        int(poolSize),
		MinIdleConns:    int(minIdleConn),
		ConnMaxIdleTime: idleTimeout,
	})

	// Enable tracing instrumentation.
	if err := redisotel.InstrumentTracing(redisClient); err != nil {
		return nil, err
	}
	// Enable metrics instrumentation.
	if err := redisotel.InstrumentMetrics(redisClient); err != nil {
		return nil, err
	}

	err := redisClient.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}

	return redisClient, nil
}

func newRedisClusterClient(c *conf.Data_Redis) (*redis.ClusterClient, error) {
	poolSize := c.GetPoolSize()
	if poolSize == 0 {
		poolSize = redisDefaultPoolSize
	}
	minIdleConn := c.MinIdleConn
	if minIdleConn <= 0 {
		minIdleConn = redisDefaultMinIdleConn
	}
	maxRetries := c.GetPoolSize()
	if maxRetries <= 0 {
		maxRetries = redisDefaultMaxRetries
	}
	idleTimeout := c.GetIdleTimeout().AsDuration()
	if idleTimeout <= 0 {
		idleTimeout = redisDefaultIdleTimeout
	}

	clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:           []string{c.Addr},
		Username:        c.GetUsername(),
		Password:        c.GetPassword(),
		MaxRetries:      int(maxRetries),
		ReadTimeout:     c.GetReadTimeout().AsDuration(),
		WriteTimeout:    c.GetWriteTimeout().AsDuration(),
		PoolSize:        int(poolSize),
		MinIdleConns:    int(minIdleConn),
		ConnMaxIdleTime: idleTimeout,
	})
	status := clusterClient.Ping(context.Background())
	if status.Err() != nil {
		return nil, errors.WithStack(status.Err())
	}

	// Enable tracing instrumentation.
	if err := redisotel.InstrumentTracing(clusterClient); err != nil {
		return nil, err
	}
	// Enable metrics instrumentation.
	if err := redisotel.InstrumentMetrics(clusterClient); err != nil {
		return nil, err
	}

	return clusterClient, nil
}
