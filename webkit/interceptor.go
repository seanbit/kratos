package webkit

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/pkg/errors"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/redis/go-redis/v9"
)

const (
	InterceptionKeyPrefix = "traffic_interception_key_%s"

	UserIdKey = "user_id"
)

var (
	once sync.Once

	currentServerName string

	interceptRedisCli redis.UniversalClient

	// interceptConf 使用 atomic.Value 保证并发安全
	// 在 goroutine 中定期更新配置，在请求处理中读取配置
	interceptConf atomic.Value // 存储 *InterceptConfig

	interceptionKey string

	// 签名配置
	signConfig atomic.Value // 存储 *SignConfig
)

// SignConfig 签名验证配置
type SignConfig struct {
	// Secret HMAC密钥，为空则禁用签名验证
	Secret string
	// SignatureLength 签名长度（hex字符数），默认8
	SignatureLength int
	// MaxTimeDrift 最大时间偏差（秒），防止重放攻击，默认300秒（5分钟）
	MaxTimeDrift int64
	// Enabled 是否启用签名验证
	Enabled bool
}

// DefaultSignConfig 返回默认签名配置
func DefaultSignConfig() *SignConfig {
	return &SignConfig{
		Secret:          "",
		SignatureLength: 8,
		MaxTimeDrift:    300,
		Enabled:         false,
	}
}

type InterceptConfig struct {
	SubRules []SubRuleConfig `json:"sub_rules"`
	Radio    int             `json:"radio"`
	Switch   bool            `json:"switch"`
}

type SubRuleConfig struct {
	Path  string `json:"path"`
	Rule  string `json:"rule"`
	Value string `json:"value"`
	Radio int    `json:"radio"`
}

// InitInterceptConfig 初始化流量拦截配置
// serverName: 服务名称
// redisCli: Redis客户端
// signCfg: 签名配置，传nil使用默认配置（禁用签名验证）
func InitInterceptConfig(serverName string, redisCli redis.UniversalClient, signCfg *SignConfig) {
	once.Do(func() {
		ctx := context.Background()
		if redisCli == nil {
			panic("init intercept redis client failed:invalid client")
		}
		currentServerName = serverName
		interceptionKey = fmt.Sprintf(InterceptionKeyPrefix, serverName)
		interceptRedisCli = redisCli

		// 初始化签名配置
		if signCfg == nil {
			signCfg = DefaultSignConfig()
		}
		if signCfg.SignatureLength <= 0 {
			signCfg.SignatureLength = 8
		}
		if signCfg.MaxTimeDrift <= 0 {
			signCfg.MaxTimeDrift = 300
		}
		signConfig.Store(signCfg)

		time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
		loadInterceptConfig(ctx)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("relaod interception config panic:%v", r)
				}
			}()
			ticker := time.NewTicker(3 * time.Second)
			for range ticker.C {
				loadInterceptConfig(ctx)
			}
			ticker.Stop()
		}()
	})
}

// UpdateSignConfig 动态更新签名配置
func UpdateSignConfig(signCfg *SignConfig) {
	if signCfg == nil {
		return
	}
	if signCfg.SignatureLength <= 0 {
		signCfg.SignatureLength = 8
	}
	if signCfg.MaxTimeDrift <= 0 {
		signCfg.MaxTimeDrift = 300
	}
	signConfig.Store(signCfg)
}

// getSignConfig 安全地获取签名配置
func getSignConfig() *SignConfig {
	if v := signConfig.Load(); v != nil {
		return v.(*SignConfig)
	}
	return DefaultSignConfig()
}

func loadInterceptConfig(ctx context.Context) {
	data, err := interceptRedisCli.Get(ctx, interceptionKey).Result()
	if err != nil {
		log.Errorf("relaod interception config failed:%v", err)
		return
	}
	var conf InterceptConfig
	err = json.Unmarshal([]byte(data), &conf)
	if err != nil {
		log.Errorf("unmarshal interception config failed:%v, data:%v", err, data)
		return
	}
	// 使用原子操作存储配置，保证并发安全
	interceptConf.Store(&conf)
}

// getInterceptConfig 安全地获取拦截配置
func getInterceptConfig() *InterceptConfig {
	if v := interceptConf.Load(); v != nil {
		return v.(*InterceptConfig)
	}
	return nil
}

// TrafficInterceptMiddleware traffic interception middleware
func TrafficInterceptMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			var path, hasSign, abnormal, block string
			defer func() {
				RecordMetricInterceptWithCtx(ctx, currentServerName, path, hasSign, abnormal, block)
				//MetricIntercept.WithLabelValues(currentServerName, path, hasSign, abnormal, block).Inc()
			}()
			// parse sign
			if tr, ok := transport.FromServerContext(ctx); ok {
				if ht, ok := tr.(khttp.Transporter); ok {
					path = ht.Request().URL.Path
					requestTime := ht.Request().Header.Get("Request-Time")
					if requestTime == "" {
						hasSign = "0"
					} else {
						hasSign = "1"
					}
					if !verifySign(ctx, requestTime) {
						abnormal = "1"
						// record traffic feature
						recordFeature(ctx, ht.Request())
						err := getInterceptStrategy(ctx, ht.Request())
						block = "0"
						if err != nil {
							block = "1"
							return nil, err
						}
					} else {
						abnormal = "0"
					}
				}
			}
			reply, err = handler(ctx, req)
			return
		}
	}
}

// verifySign 验证请求签名
// 请求头格式: Request-Time: {timestamp}.{signature}
// 其中 timestamp 为毫秒时间戳，signature 为 HMAC-SHA256(secret, timestamp) 的前N位hex
func verifySign(ctx context.Context, requestTime string) bool {
	cfg := getSignConfig()

	// 未启用签名验证，直接返回true
	if !cfg.Enabled || cfg.Secret == "" {
		return true
	}

	// 解析 timestamp.signature 格式
	parts := strings.SplitN(requestTime, ".", 2)
	if len(parts) != 2 {
		log.Context(ctx).Warnw(
			"msg", "invalid request_time format, expected: timestamp.signature",
			"request_time", requestTime,
		)
		return false
	}

	timestampStr := parts[0]
	signature := parts[1]

	// 验证时间戳格式
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		log.Context(ctx).Warnw(
			"msg", "invalid timestamp",
			"timestamp", timestampStr,
		)
		return false
	}

	// 检查时间偏差（防止重放攻击）
	now := time.Now().UnixMilli()
	drift := now - timestamp
	if drift < 0 {
		drift = -drift
	}
	maxDriftMs := cfg.MaxTimeDrift * 1000
	if drift > maxDriftMs {
		log.Context(ctx).Warnw(
			"msg", "request_time drift too large",
			"timestamp", timestamp,
			"now", now,
			"drift_ms", drift,
			"max_drift_ms", maxDriftMs,
		)
		return false
	}

	// 验证 HMAC-SHA256 签名
	expectedSig := generateSignature(timestampStr, cfg.Secret, cfg.SignatureLength)
	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		log.Context(ctx).Warnw(
			"msg", "signature mismatch",
			"expected_length", cfg.SignatureLength,
		)
		return false
	}

	return true
}

// generateSignature 生成 HMAC-SHA256 签名
// 返回签名的前 length 位十六进制字符
func generateSignature(data, secret string, length int) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	fullSig := hex.EncodeToString(h.Sum(nil))
	if length > len(fullSig) {
		length = len(fullSig)
	}
	return fullSig[:length]
}

// GenerateRequestTime 生成带签名的请求时间（供客户端使用）
// 返回格式: {timestamp}.{signature}
func GenerateRequestTime(secret string, signatureLength int) string {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signature := generateSignature(timestamp, secret, signatureLength)
	return timestamp + "." + signature
}

// sensitiveHeaders 敏感请求头列表（小写）
var sensitiveHeaders = map[string]bool{
	"authorization":       true,
	"cookie":              true,
	"set-cookie":          true,
	"x-api-key":           true,
	"x-auth-token":        true,
	"x-access-token":      true,
	"x-refresh-token":     true,
	"proxy-authorization": true,
	"www-authenticate":    true,
}

// sanitizeHeaders 过滤敏感请求头
func sanitizeHeaders(headers http.Header) map[string]string {
	sanitized := make(map[string]string, len(headers))
	for key, values := range headers {
		lowerKey := strings.ToLower(key)
		if sensitiveHeaders[lowerKey] {
			sanitized[key] = "***REDACTED***"
		} else {
			sanitized[key] = strings.Join(values, ", ")
		}
	}
	return sanitized
}

func recordFeature(ctx context.Context, req *http.Request) {
	path := req.URL.Path
	referer := req.Header.Get("Referer")
	ua := req.Header.Get("User-Agent")
	uid, _ := UserIdFromContext(ctx)
	log.Context(ctx).Warnw(
		"msg", "abnormal traffic feature",
		"path", path,
		"referer", referer,
		"ua", ua,
		"uid", uid,
		"header", sanitizeHeaders(req.Header),
	)
}

func getInterceptStrategy(ctx context.Context, req *http.Request) error {
	path := req.URL.Path
	referer := req.Header.Get("Referer")
	ua := req.Header.Get("User-Agent")
	uid, _ := UserIdFromContext(ctx)

	// 使用原子操作安全地获取配置
	conf := getInterceptConfig()
	if conf == nil || !conf.Switch {
		return nil
	}
	// global rule
	err := getStrategyByRadio(conf.Radio)
	if err != nil {
		return err
	}
	for _, subRule := range conf.SubRules {
		if !matchRule(subRule, path, referer, ua, uid) {
			continue
		}
		err := getStrategyByRadio(subRule.Radio)
		if err != nil {
			return err
		}
	}
	return nil
}

func matchRule(rule SubRuleConfig, path, referer, ua, uid string) bool {
	switch rule.Rule {
	case "*":
		return true
	case "path":
		return inSliceStr(path, strings.Split(rule.Value, ","))
	case "referer":
		return sliceContain(referer, strings.Split(rule.Value, ","))
	case "ua":
		return inSliceStr(ua, strings.Split(rule.Value, ","))
	case "uid":
		return inSliceStr(uid, strings.Split(rule.Value, ","))
	default:
		return false
	}
}

func getStrategyByRadio(radio int) error {
	if radio == -1 {
		return nil
	}
	if radio == 0 {
		// TODO send lark
		return nil
	}
	if radio > 0 && radio <= 100 {
		if rand.Intn(200)%100 < radio {
			return errors.New("invalid reqeust")
		}
	}
	return nil
}

func inSliceStr(source string, target []string) bool {
	for _, item := range target {
		if item == source {
			return true
		}
	}
	return false
}

func sliceContain(source string, target []string) bool {
	for _, item := range target {
		if strings.Contains(source, item) {
			return true
		}
	}
	return false
}
