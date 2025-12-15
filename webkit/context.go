package webkit

import (
	"context"
	"net"
	"strings"

	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	"go.opentelemetry.io/otel/trace"
)

type _AppInfoKey struct{}

var (
	appInfoKey  = _AppInfoKey{}
	userInfoKey = "user_info"
)

type App struct {
	AppID          string   `json:"app_id"`
	APIKey         string   `json:"api_key"`
	IsAdmin        bool     `json:"is_admin"`
	AppName        string   `json:"app_name"`
	AppLogoURL     string   `json:"app_logo_url"`
	Permissions    int64    `json:"permissions"`
	AllowedDomains []string `json:"allowed_domains"`
}

type UserInfo struct {
	UserId        string `json:"user_id"`
	Username      string `json:"username"`
	UserType      string `json:"user_type"`
	WalletAddress string `json:"wallet_address"`
}

func NewAppInfoContext(ctx context.Context, info *App) context.Context {
	return context.WithValue(ctx, appInfoKey, info)
}

func NewUserInfoContext(ctx context.Context, info *UserInfo) context.Context {
	return context.WithValue(ctx, userInfoKey, info)
}

func UserInfoFromContext(ctx context.Context) *UserInfo {
	val := ctx.Value(userInfoKey)
	userInfo, ok := val.(*UserInfo)
	if !ok || userInfo == nil || userInfo.UserId == "" {
		return nil
	}
	return userInfo
}

func UserIdFromContext(ctx context.Context) (string, bool) {
	userInfo := UserInfoFromContext(ctx)
	if userInfo == nil {
		return "", false
	}
	return userInfo.UserId, true
}

func GetRealIP(ctx context.Context) string {
	transPort, ok := transport.FromServerContext(ctx)
	if !ok {
		return ""
	}
	if transPort.Kind() != transport.KindHTTP {
		return ""
	}

	info, ok := transPort.(*http.Transport)
	if !ok {
		return ""
	}

	r := info.Request()
	ip := r.Header.Get("X-Real-IP")
	if ip != "" && !strings.HasPrefix(ip, "10.") {
		return ip
	}

	ip = r.Header.Get("X-Forwarded-For")
	for _, i := range strings.Split(ip, ",") {
		if net.ParseIP(i) != nil && !strings.HasPrefix(i, "10.") {
			return i
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return ""
	}

	if net.ParseIP(ip) != nil && !strings.HasPrefix(ip, "10.") {
		return ip
	}
	return ""
}

func GetTraceID(ctx context.Context) (traceID string) {
	if _, ok := transport.FromServerContext(ctx); ok {
		if span := trace.SpanContextFromContext(ctx); span.HasTraceID() {
			return span.TraceID().String()
		}
	}
	return
}

func GetPlatformFromHeader(ctx context.Context) string {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return "web"
	}
	platform := tr.RequestHeader().Get("platform")
	if platform == "" {
		platform = "web"
	}
	return platform
}

func GetOperationFromContext(ctx context.Context) (operation string) {
	if info, ok := transport.FromServerContext(ctx); ok {
		operation = info.Operation()
	}
	return operation
}
