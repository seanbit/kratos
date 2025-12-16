package data

import (
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewAlarmMessageRepo, NewAlarm,
	NewAuthRepo, NewAuthLogRepo,
	NewGeoIP,
)
