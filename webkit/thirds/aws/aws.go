package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/translate"
)

var client *Client

type Client struct {
	region          string
	sesClient       *sesv2.Client
	translateClient *translate.Client
}

func GetClient() *Client {
	return client
}

func GetTranslateClient() *translate.Client {
	return client.translateClient
}

type Config struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
}

func Init(cf *Config) error {
	credProvider := credentials.NewStaticCredentialsProvider(cf.AccessKey, cf.SecretKey, "")
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cf.Region),
		config.WithCredentialsProvider(credProvider),
	)
	if err != nil {
		return err
	}

	client = &Client{
		region:          cf.Region,
		sesClient:       sesv2.NewFromConfig(cfg),
		translateClient: translate.NewFromConfig(cfg),
	}

	return nil
}
