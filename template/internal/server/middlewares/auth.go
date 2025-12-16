package middlewares

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/seanbit/kratos/template/internal/biz"
	"github.com/seanbit/kratos/webkit"
	"google.golang.org/grpc/metadata"
)

type IUserInfoService interface {
	GetUserInfoByAuthToken(ctx context.Context, authToken string) (userInfo *webkit.UserInfo, err error)
}
type UserAuth struct {
	userInfoServ IUserInfoService
}

func NewUserAuth(userInfoServ *biz.Auth) *UserAuth {
	return &UserAuth{userInfoServ: userInfoServ}
}

func (mw *UserAuth) Build() middleware.Middleware {
	handler := func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			var (
				jwtToken string
				platform string
			)

			if md, ok := metadata.FromIncomingContext(ctx); ok {
				jwtToken = md.Get("x-md-global-jwt-key-gen")[0]
			} else if tr, ok := transport.FromServerContext(ctx); ok {
				jwtToken, err = webkit.FromAuthHeader(tr)
				if err != nil {
					return nil, err
				}

				platform = tr.RequestHeader().Get("platform")
				if platform == "" {
					platform = "web"
				}
			} else {
				// 缺少可认证的token，返回错误
				return nil, webkit.ErrAuthFail
			}

			userInfo, err := mw.userInfoServ.GetUserInfoByAuthToken(ctx, jwtToken)
			if err != nil {
				return nil, errors.New(4200, "tips: jwt-key-gen token verify failed", fmt.Sprintf("Get user info failed: %v", err))
			}
			ctx = webkit.NewUserInfoContext(ctx, userInfo)
			reply, err = handler(ctx, req)
			return
		}
	}
	return selector.Server(handler).Match(NewWhiteListMatcher()).Build()
}

// NewWhiteListMatcher jwt白名单，在名单中的路由不用校验jwt
func NewWhiteListMatcher() selector.MatchFunc {
	whiteList := make(map[string]struct{})
	// key的格式依据proto定义的: /package.service/rpcName, 例如: /carv_api.Auth/CheckSignupStatus
	whiteList["/probe.Probe/healthy"] = struct{}{}
	whiteList["/probe.Probe/ready"] = struct{}{}
	whiteList["/auth.Auth/GetLoginSignatureText"] = struct{}{}
	whiteList["/auth.Auth/LoginByWallet"] = struct{}{}

	return func(ctx context.Context, operation string) bool {
		//log.Context(ctx).Infof("whiteList operation: %v", operation)
		if _, ok := whiteList[operation]; ok {
			return true
		}
		return false
	}
}
