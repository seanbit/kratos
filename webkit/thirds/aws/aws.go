package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/translate"
)

// Deprecated: use NewClient instead of the global client pattern.
var client *Client

type Client struct {
	region          string
	sesClient       *sesv2.Client
	translateClient *translate.Client
}

// GetClient returns the global client. Deprecated: prefer NewClient for DI-friendly usage.
func GetClient() *Client {
	return client
}

// GetTranslateClient returns the global translate client. Deprecated: prefer NewClient.
func GetTranslateClient() *translate.Client {
	return client.translateClient
}

func (c *Client) SESClient() *sesv2.Client       { return c.sesClient }
func (c *Client) TranslateClient() *translate.Client { return c.translateClient }

type Config struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
}

// Init initializes the global client. Deprecated: prefer NewClient.
func Init(cf *Config) error {
	c, err := NewClient(cf)
	if err != nil {
		return err
	}
	client = c
	return nil
}

// NewClient creates a new AWS client instance, suitable for dependency injection.
func NewClient(cf *Config) (*Client, error) {
	credProvider := credentials.NewStaticCredentialsProvider(cf.AccessKey, cf.SecretKey, "")
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cf.Region),
		config.WithCredentialsProvider(credProvider),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		region:          cf.Region,
		sesClient:       sesv2.NewFromConfig(cfg),
		translateClient: translate.NewFromConfig(cfg),
	}, nil
}
