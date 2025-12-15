package webkit

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"cosmossdk.io/errors"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/oschwald/geoip2-golang"
	"github.com/tencentyun/cos-go-sdk-v5"
)

type GeoSource interface {
	GetGeoIpFile(context.Context) (io.Reader, error)
}

type CosGeoSource struct {
	fileName  string
	cosClient *cos.Client
}

func NewCosGeoSource(cosClient *cos.Client, fileName string) *CosGeoSource {
	return &CosGeoSource{cosClient: cosClient, fileName: fileName}
}

func (s *CosGeoSource) GetGeoIpFile(ctx context.Context) (io.Reader, error) {
	resp, err := s.cosClient.Object.Get(ctx, s.fileName, nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

type S3GeoSource struct {
	bucketName string
	fileName   string
	s3Client   *s3.Client
}

func NewS3GeoSource(s3Client *s3.Client, bucketName, fileName string) *S3GeoSource {
	return &S3GeoSource{s3Client: s3Client, bucketName: bucketName, fileName: fileName}
}

func (s *S3GeoSource) GetGeoIpFile(ctx context.Context) (io.Reader, error) {
	objectOutput, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(s.fileName),
	})
	if err != nil {
		return nil, err
	}
	return objectOutput.Body, nil
}

type GeoIP struct {
	source GeoSource
	cli    *geoip2.Reader
}

func NewGeoIP(source GeoSource) (*GeoIP, error) {
	if source == nil {
		return nil, fmt.Errorf("[NewGeoIP] source is null")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	dbReader, err := source.GetGeoIpFile(ctx)
	if err != nil {
		log.Context(ctx).Errorf("[NewGeoIP] get file failed:%v", err)
		return nil, err
	}

	tempFile, err := os.CreateTemp("", "geo_ip_tmp")
	if err != nil {
		return nil, errors.Wrap(err, "[NewGeoIP] create file failed")
	}
	_, err = io.Copy(tempFile, dbReader)
	if err != nil {
		return nil, errors.Wrap(err, "[NewGeoIP] copy file failed")
	}
	err = tempFile.Sync()
	if err != nil {
		return nil, errors.Wrap(err, "[NewGeoIP] sync file failed")
	}

	geoIp := &GeoIP{source: source}
	geoIp.cli, err = geoip2.Open(tempFile.Name())
	if err != nil {
		return nil, errors.Wrap(err, "[NewGeoIP] open file failed")
	}
	return geoIp, nil
}

func (geo *GeoIP) GetCountryByIp(ipStr string) (*geoip2.Country, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("parse ip nil")
	}
	if geo.cli == nil {
		return nil, fmt.Errorf("geo ip client is nil")
	}
	return geo.cli.Country(ip)
}
