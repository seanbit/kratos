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

	// 验证配置
	if err := ValidateConfig(&bc); err != nil {
		panic(fmt.Sprintf("config validation failed: %v", err))
	}

	SetConfig(&bc)

	return
}

// ConfigValidationError 配置验证错误
type ConfigValidationError struct {
	Field   string
	Message string
}

func (e *ConfigValidationError) Error() string {
	return fmt.Sprintf("config validation error: %s - %s", e.Field, e.Message)
}

// ValidateConfig 验证配置完整性和有效性
// 在服务启动前调用，确保必要配置项存在
func ValidateConfig(bc *conf.Bootstrap) error {
	var errors []string

	// 验证服务名称
	if bc.Name == "" {
		errors = append(errors, "name: service name is required")
	}

	// 验证 Server 配置
	if bc.Server == nil {
		errors = append(errors, "server: server configuration is required")
	} else {
		if bc.Server.Http == nil || bc.Server.Http.Addr == "" {
			errors = append(errors, "server.http.addr: HTTP server address is required")
		}
	}

	// 验证 Data 配置
	if bc.Data == nil {
		errors = append(errors, "data: data configuration is required")
	} else {
		// 验证数据库配置
		if bc.Data.Database == nil {
			errors = append(errors, "data.database: database configuration is required")
		} else {
			if bc.Data.Database.Driver == "" {
				errors = append(errors, "data.database.driver: database driver is required")
			}
		}

		// 验证 Redis 配置
		if bc.Data.Redis == nil {
			errors = append(errors, "data.redis: redis configuration is required")
		} else {
			if bc.Data.Redis.Addr == "" {
				errors = append(errors, "data.redis.addr: redis address is required")
			}
		}
	}

	// 验证 Auth 配置
	if bc.Auth == nil {
		errors = append(errors, "auth: auth configuration is required")
	} else {
		if bc.Auth.JwtKey_25519 == "" {
			errors = append(errors, "auth.jwt_key_25519: JWT key is required")
		}
		if bc.Auth.LoginExpires == nil || bc.Auth.LoginExpires.AsDuration() <= 0 {
			errors = append(errors, "auth.login_expires: login expiration duration is required")
		}
	}

	// 验证 Alarm 配置（如果启用）
	if bc.Alarm != nil && !bc.Alarm.DryRun {
		if bc.Alarm.DefaultPlatform == "" {
			errors = append(errors, "alarm.default_platform: default platform is required when alarm is enabled")
		}
		if len(bc.Alarm.WebHooks) == 0 {
			errors = append(errors, "alarm.web_hooks: at least one webhook is required when alarm is enabled")
		} else if bc.Alarm.DefaultPlatform != "" {
			if _, ok := bc.Alarm.WebHooks[bc.Alarm.DefaultPlatform]; !ok {
				errors = append(errors, fmt.Sprintf("alarm.web_hooks: webhook for default platform '%s' is not configured", bc.Alarm.DefaultPlatform))
			}
		}
	}

	// 验证 Tracing 配置（如果配置了 host）
	if bc.Tracing != nil && bc.Tracing.Host != "" {
		if bc.Tracing.Port <= 0 {
			errors = append(errors, "tracing.port: tracing port must be positive when host is configured")
		}
		if bc.Tracing.Type == "" {
			errors = append(errors, "tracing.type: tracing type is required when host is configured")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration errors:\n  - %s", joinStrings(errors, "\n  - "))
	}

	return nil
}

// joinStrings 连接字符串切片
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
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
