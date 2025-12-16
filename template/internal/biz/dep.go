package biz

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type IAlarmRepo interface {
	SendBizMessage(ctx context.Context, title, info string)
	SendMessage(ctx context.Context, platform, title, info string)
}

type GeoCountry struct {
	IsoCode string
}
type IGeoIp interface {
	GetCountryFromIp(ctx context.Context, ip string) (*GeoCountry, error)
}

type NavInfo struct {
	CurrentValue   decimal.Decimal `json:"current_value"`
	Rebalancing    bool            `json:"rebalancing"`
	RebalancedTime time.Time       `json:"rebalanced_time"`
}

//go:generate mockgen -source=dep.go -destination=./mocks/nav_service.go -package=mocks
type INavService interface {
	GetNav() (*NavInfo, error)
}
