package webkit

import (
	"context"
	"fmt"
	nethttp "net/http"
	"net/url"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/pkg/errors"
)

const (
	Origin           = "Origin"
	Referer          = "Referer"
	AllowCredentials = "Access-Control-Allow-Credentials"
	AllowOrigin      = "Access-Control-Allow-Origin"
	AllowMethods     = "Access-Control-Allow-Methods"
	AllowHeaders     = "Access-Control-Allow-Headers"
	ExposeHeaders    = "Access-Control-Expose-Headers"
	SecurityHeaders  = "Strict-Transport-Security"
)

func DealWithHeader(tr transport.Transporter, allowDomains []string) error {
	origin := tr.RequestHeader().Get(Origin)
	if origin == "" {
		origin = tr.RequestHeader().Get(Referer)
	}
	if origin != "" {
		u, err := url.Parse(origin)
		if err != nil {
			return err
		}

		var allowDomain bool
		for _, v := range allowDomains {
			if u.Host == v || strings.HasSuffix(u.Host, fmt.Sprintf(".%s", v)) {
				allowDomain = true
				break
			}
		}
		if !allowDomain {
			return fmt.Errorf("%s is not allowed in %s", origin, tr.RequestHeader().Get(Origin))
		}
	}

	if tr.ReplyHeader().Get(AllowOrigin) == "" {
		if origin != "" {
			tr.ReplyHeader().Set(AllowOrigin, origin)
		}
	}
	if tr.ReplyHeader().Get(AllowMethods) == "" {
		tr.ReplyHeader().Set(AllowMethods, "GET,POST,OPTIONS,PUT,PATCH,DELETE")
	}
	if tr.ReplyHeader().Get(AllowCredentials) == "" {
		tr.ReplyHeader().Set(AllowCredentials, "true")
	}
	if tr.ReplyHeader().Get(ExposeHeaders) == "" {
		tr.ReplyHeader().Set(ExposeHeaders, "Content-Length,X-Token")
	}
	if tr.ReplyHeader().Get(AllowHeaders) == "" {
		tr.ReplyHeader().Set(AllowHeaders, "Content-Type,Recaptcha-Token,X-Token,Platform"+
			"X-Requested-With,Access-Control-Allow-Credentials,User-Agent,Content-Length,Authorization,Locale,Source")
	}
	if tr.ReplyHeader().Get(SecurityHeaders) == "" {
		tr.ReplyHeader().Set(SecurityHeaders, "max-age=31536000")
	}

	return nil
}

// MiddlewareCors 设置跨域请求头
func MiddlewareCors(allowDomains []string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				if err := DealWithHeader(tr, allowDomains); err != nil {
					return nil, err
				}

				return handler(ctx, req)
			}
			return nil, errors.New("can't found transport from ctx")
		}
	}
}

type AllowMethodsCORS struct {
	h nethttp.Handler
}

func (amc *AllowMethodsCORS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	}
	amc.h.ServeHTTP(w, r)
}

func AllowMethodsCORSFunc() func(nethttp.Handler) nethttp.Handler {
	return func(h nethttp.Handler) nethttp.Handler {
		amc := &AllowMethodsCORS{}
		amc.h = h
		return amc
	}
}
