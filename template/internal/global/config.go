package global

import (
	"fmt"
	"os"
	"sync"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/seanbit/kratos/template/internal/conf"
	"github.com/seanbit/kratos/webkit/thirds/aws"
	"gopkg.in/yaml.v3"
)

var (
	cnf  *conf.Bootstrap
	lock sync.Mutex
)

func SetConfig(c *conf.Bootstrap) {
	lock.Lock()
	defer lock.Unlock()
	cnf = c
	os.Setenv("ENV", cnf.Env.String())
}

func GetConfig() *conf.Bootstrap {
	lock.Lock()
	defer lock.Unlock()

	return cnf
}

func InitConfig(confSrc, confPath, secretFile string) (clean func()) {
	// set local env
	if secretFile != "" {
		if err := LoadSecretFromFile(confPath, secretFile); err != nil {
			panic(fmt.Sprintf("load secret from file failed: %v", err))
		}
	}

	var src config.Source
	switch confSrc {
	case "file":
		src = file.NewSource(confPath)
	case "aws-appconfig":
		src = aws.NewConfigSource()
	default:
		panic("unknown config source")
	}
	// 加载配置
	c := config.New(
		config.WithSource(src),
		config.WithDecoder(envDecoder),
	)
	clean = func() { _ = c.Close() }

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}
	SetConfig(&bc)

	return
}

func envDecoder(kv *config.KeyValue, v map[string]interface{}) error {
	// 用 yaml.v3 解码
	if err := yaml.Unmarshal(kv.Value, &v); err != nil {
		return err
	}
	// 遍历替换所有字符串里的 ${ENV_VAR}
	var replaceEnv func(interface{}) interface{}
	replaceEnv = func(val interface{}) interface{} {
		switch vv := val.(type) {
		case string:
			return os.ExpandEnv(vv)
		case map[string]interface{}:
			for k, v2 := range vv {
				vv[k] = replaceEnv(v2)
			}
			return vv
		case []interface{}:
			for i, v2 := range vv {
				vv[i] = replaceEnv(v2)
			}
			return vv
		default:
			return val
		}
	}
	for k, v2 := range v {
		v[k] = replaceEnv(v2)
	}
	return nil
}
