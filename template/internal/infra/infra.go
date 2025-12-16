package infra

import (
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(
	NewDataProvider,
	wire.Bind(new(PostgresProvider), new(*DataProvider)),
	wire.Bind(new(RedisProvider), new(*DataProvider)),
	NewS3Client,
)
