package webkit

import (
	"context"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/transport"
)

type _JwtTokenKey struct{}

var (
	ErrAuthFail     = errors.New(4200, "tips: jwt-key-gen token failed", "")
	ErrUnknowClaims = errors.New(4200, "tips: jwt-key-gen token unknow claims", "")
	ErrExpired      = errors.New(4200, "tips: jwt-key-gen is expired", "")
	ErrUnlogin      = errors.New(4200, "tips: not login", "")
	ErrInvalidAppID = errors.New(4200, "tips: appid is invalid", "")
	ErrNotMerge     = errors.New(4200, "Please update your CARV app to the latest version, or login again", "Please update your CARV app to the latest version, or login again")
	ErrServer       = errors.New(5000, "tips: server error", "")
	jwtTokenKey     = _JwtTokenKey{}
)

const (
	Authorization      = "Authorization"
	defaultAppIDHeader = "x-app-id"
	Expired            = 60 * 24 * 30 // 分钟
	LoginSessionTime   = time.Hour * 24 * 7
)

//==============

// FromAuthHeader is a "TokenExtractor" that takes a give context and extracts
// the JWT token from the Authorization header.
func FromAuthHeader(tr transport.Transporter) (string, error) {
	authHeader := tr.RequestHeader().Get(Authorization)
	if authHeader == "" {
		return "", nil // No error, just no token
	}

	// TODO: Make this a bit more robust, parsing-wise
	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) == 2 && strings.ToLower(authHeaderParts[0]) == "bearer" {
		return authHeaderParts[1], nil
	} else {
		return authHeader, nil
	}
}

func GetHeader(ctx context.Context, key string) string {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return ""
	}

	return tr.RequestHeader().Get(key)
}

func NewJwtInfoContext(ctx context.Context, claims any) context.Context {
	return context.WithValue(ctx, jwtTokenKey, claims)
}
