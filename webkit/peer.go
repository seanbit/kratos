package webkit

import (
	"context"
	"net"
	"strings"

	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func GetClientIP(ctx context.Context) string {
	if transPort, ok := transport.FromServerContext(ctx); ok {
		//http调用还是grpc调用
		if transPort.Kind() == transport.KindHTTP {
			if info, ok := transPort.(*http.Transport); ok {
				// http打印信息
				request := info.Request()
				clientIP, _, _ := net.SplitHostPort(request.RemoteAddr)
				return clientIP
			}
		} else {
			// grpc打印信息
			if _, ok := transPort.(*grpc.Transport); ok {
				if pr, ok := peer.FromContext(ctx); ok {
					clientIP, _, _ := net.SplitHostPort(pr.Addr.String())
					return clientIP
				}
			}
		}
	}

	return ""
}

func GetUserAgent(ctx context.Context) string {
	if transPort, ok := transport.FromServerContext(ctx); ok {
		//http调用还是grpc调用
		if transPort.Kind() == transport.KindHTTP {
			if info, ok := transPort.(*http.Transport); ok {
				// http打印信息
				request := info.Request()
				userAgent := request.UserAgent()
				return userAgent
			}
		} else {
			// grpc打印信息
			if _, ok := transPort.(*grpc.Transport); ok {
				md, _ := metadata.FromIncomingContext(ctx)
				userAgent := md.Get("user-agent")
				return strings.Join(userAgent, " ")
			}
		}
	}

	return ""
}
