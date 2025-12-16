package infra

import (
	"context"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/seanbit/kratos/template/internal/conf"
)

func NewS3Client(config *conf.S3) *s3.Client {
	credProvider := credentials.NewStaticCredentialsProvider(config.AccessKey, config.SecretKey, "")
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion(config.Region),
		awsConfig.WithCredentialsProvider(credProvider),
	)
	if err != nil {
		log.Errorf("failed to load aws config: %v", err)
		panic(err)
	}
	return s3.NewFromConfig(cfg)
}
