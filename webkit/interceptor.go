package webkit

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
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

	interceptConf *InterceptConfig

	interceptionKey string

	// sign
	signNum int   = 1
	factor1 int64 = 1
	factor2 int64 = 0
)

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

func InitInterceptConfig(serverName string, redis redis.UniversalClient) {
	once.Do(func() {
		ctx := context.Background()
		if redis == nil {
			panic("init intercept redis client failed:invalid client")
		}
		currentServerName = serverName
		interceptionKey = fmt.Sprintf(InterceptionKeyPrefix, serverName)
		interceptRedisCli = redis
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
	interceptConf = &conf
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
				if ht, ok := tr.(http.Transporter); ok {
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

func verifySign(ctx context.Context, requestTime string) bool {
	if len(requestTime) < 13 {
		log.Context(ctx).Warnw(
			"request_time", requestTime,
			"msg", "invalid request_time",
		)
		return false
	}
	pre := requestTime[0 : len(requestTime)-signNum]
	preNum, _ := strconv.ParseInt(pre[len(pre)-signNum:], 10, 64)
	tailNum, _ := strconv.ParseInt(requestTime[len(requestTime)-signNum:], 10, 64)
	return (factor1*preNum+factor2)%pow(10, signNum) == tailNum
}

func pow(b int64, p int) int64 {
	var ret int64 = 1
	for i := 0; i < p; i++ {
		ret *= b
	}
	return ret
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
		"header", req.Header,
	)
}

func getInterceptStrategy(ctx context.Context, req *http.Request) error {
	path := req.URL.Path
	referer := req.Header.Get("Referer")
	ua := req.Header.Get("User-Agent")
	uid, _ := UserIdFromContext(ctx)

	if interceptConf == nil || !interceptConf.Switch {
		return nil
	}
	// global rule
	err := getStrategyByRadio(interceptConf.Radio)
	if err != nil {
		return err
	}
	for _, subRule := range interceptConf.SubRules {
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
