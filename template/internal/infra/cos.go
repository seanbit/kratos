package infra

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/seanbit/kratos/template/internal/conf"
	"github.com/tencentyun/cos-go-sdk-v5"
)

func NewCosClient(config *conf.Cos) *cos.Client {
	bucketUrl, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.Bucket, config.Region))
	if err != nil {
		log.Errorf("cos bucket url parse err:%v", err)
		panic(err)
	}
	return cos.NewClient(&cos.BaseURL{BucketURL: bucketUrl}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.SecretId,
			SecretKey: config.SecretKey,
		},
	})
}
