package aws

import (
	"context"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	kconfig "github.com/go-kratos/kratos/v2/config"
	"github.com/pkg/errors"
)

type AwsConfigMeta struct {
	Region          string `env:"AWS_REGION,required"`
	ApplicationId   string `env:"AWS_APPLICATION_ID,required"`
	Environment     string `env:"AWS_ENVIRONMENT,required"`
	ConfigurationId string `env:"AWS_CONFIGURATION_ID,required"`
	ClientId        string `env:"AWS_CLIENT_ID"`
	EnableWatcher   bool   `env:"AWS_ENABLE_WATCHER,default=false"`
	WatcherInterval int    `env:"AWS_WATCHER_INTERVAL"`
}

type awsConfigure struct {
	client  *appconfig.Client
	metaCnf *AwsConfigMeta
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewConfigSource() kconfig.Source {
	metaCnf, err := loadAwsConfigMetaFromEnv()
	if err != nil {
		panic(err)
	}
	cli, err := createClient(metaCnf)
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())

	return &awsConfigure{
		client:  cli,
		metaCnf: metaCnf,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func createClient(conf *AwsConfigMeta) (*appconfig.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}
	cfg.Region = conf.Region
	c := appconfig.NewFromConfig(cfg)
	return c, nil
}

func (a *awsConfigure) Load() ([]*kconfig.KeyValue, error) {
	input := &appconfig.GetConfigurationInput{
		Application:   awssdk.String(a.metaCnf.ApplicationId),
		Environment:   awssdk.String(a.metaCnf.Environment),
		Configuration: awssdk.String(a.metaCnf.ConfigurationId),
		ClientId:      awssdk.String(a.metaCnf.ClientId),
	}

	resp, err := a.client.GetConfiguration(context.Background(), input)
	if err != nil {
		return nil, err
	}

	kv := &kconfig.KeyValue{
		Key:    a.metaCnf.ApplicationId,
		Format: "yaml",
		Value:  resp.Content,
	}

	return []*kconfig.KeyValue{kv}, nil
}

func (a *awsConfigure) Watch() (kconfig.Watcher, error) {
	return a, nil
}

func (a *awsConfigure) Next() ([]*kconfig.KeyValue, error) {
	if !a.metaCnf.EnableWatcher {
		// 不支持watch, 阻塞住
		<-a.ctx.Done()
		return nil, nil
	}

	// 最少90秒
	if a.metaCnf.WatcherInterval < 90 {
		a.metaCnf.WatcherInterval = 90
	}
	t := time.NewTimer(time.Duration(a.metaCnf.WatcherInterval) * time.Second)
	defer t.Stop()

	select {
	case <-t.C:
		return a.Load()
	case <-a.ctx.Done():
		return nil, nil
	}
}

func (a *awsConfigure) Stop() error {
	a.cancel()
	return nil
}

// ---
func loadAwsConfigMetaFromEnv() (*AwsConfigMeta, error) {
	awsConf := new(AwsConfigMeta)

	rv := reflect.ValueOf(awsConf).Elem()
	for i := 0; i < rv.NumField(); i++ {
		f := rv.Type().Field(i)
		envOpts := strings.Split(f.Tag.Get("env"), ",")
		envName, tag := envOpts[0], ""
		if len(envOpts) >= 2 {
			tag = envOpts[1]
		}
		envVal := os.Getenv(envName)

		if envVal == "" && tag == "required" {
			return nil, errors.New("missing aws env: " + envName)
		} else if envVal != "" {
			field := rv.Field(i)
			typ := field.Type().Kind()
			if typ == reflect.Bool {
				if field.CanSet() {
					field.SetBool(envVal == "true")
				}
			} else if typ == reflect.Int {
				if field.CanSet() {
					intVal, err := strconv.Atoi(envVal)
					if err != nil {
						intVal = 0
					}
					field.SetInt(int64(intVal))
				}
			} else {
				if field.CanSet() {
					field.SetString(envVal)
				}
			}
		}
	}

	if awsConf.EnableWatcher && awsConf.WatcherInterval == 0 {
		awsConf.WatcherInterval = 30
	}
	if awsConf.ClientId == "" {
		cid, _ := getClientId()
		awsConf.ClientId = cid
	}

	return awsConf, nil
}

func getClientId() (string, error) {
	// 从 /proc/self/cgroup 中读取容器 ID
	content, err := os.ReadFile("/etc/hostname")
	if err != nil {
		return "no_valid_client_id", err
	}

	return strings.TrimSpace(string(content)), nil
}
