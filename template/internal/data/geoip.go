package data

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
	"github.com/seanbit/kratos/template/internal/biz"
	"github.com/seanbit/kratos/template/internal/conf"
	"github.com/seanbit/kratos/template/internal/global"
	"github.com/seanbit/kratos/webkit"
)

type geoIp struct {
	s3Client *s3.Client
	config   *conf.GeoIp
	source   *webkit.GeoIP
}

func NewGeoIP(s3Client *s3.Client, config *conf.GeoIp) (biz.IGeoIp, error) {
	s3GeoSource := webkit.NewS3GeoSource(s3Client, config.FileBucket, config.FileKey)
	repo := &geoIp{
		s3Client: s3Client,
		config:   config,
	}
	if global.GetEnv() != conf.Env_LOCAL.String() {
		var err error
		repo.source, err = webkit.NewGeoIP(s3GeoSource)
		if err != nil {
			return nil, errors.Wrap(err, "NewGeoIP")
		}
	}
	return repo, nil
}

func (repo *geoIp) GetCountryFromIp(ctx context.Context, ip string) (*biz.GeoCountry, error) {
	if global.GetEnv() == conf.Env_LOCAL.String() {
		return &biz.GeoCountry{}, nil
	}
	result, err := repo.source.GetCountryByIp(ip)
	if err != nil {
		return nil, errors.Wrap(err, "GetCountryFromIp")
	}
	return &biz.GeoCountry{IsoCode: result.Country.IsoCode}, nil
}
