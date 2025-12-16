package infra

import (
	"github.com/seanbit/kratos/template/internal/conf"
)

type DataProvider struct {
	PostgresProvider
	RedisProvider
}

// NewDataProvider .
func NewDataProvider(c *conf.Data) (*DataProvider, func(), error) {
	var (
		provider = &DataProvider{}
		err      error
	)
	defer func() {
		if err != nil {
			provider.Close()
		}
	}()

	provider.PostgresProvider, _, err = NewPostgresProvider(c.GetDatabase())
	if err != nil {
		return nil, nil, err
	}
	provider.RedisProvider, _, err = NewRedisProvider(c.GetRedis())
	if err != nil {
		return nil, nil, err
	}
	return provider, provider.Close, nil
}

func (provider *DataProvider) Close() {
	if provider.PostgresProvider != nil {
		provider.PostgresProvider.Close()
	}
	if provider.RedisProvider != nil {
		provider.RedisProvider.Close()
	}
}
