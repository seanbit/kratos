// transport/asynq/transport.go
package asynq

import (
	"github.com/go-kratos/kratos/v2/transport"
)

const KindASYNQ transport.Kind = "asynq"

var _ transport.Transporter = (*Server)(nil)

// Kind 返回传输类型
func (s *Server) Kind() transport.Kind {
	return KindASYNQ
}

// Endpoint 返回终端地址
func (s *Server) Endpoint() string {
	if s.config == nil {
		return ""
	}
	return s.config.RedisURI
}

// Operation 返回操作名称
func (s *Server) Operation() string {
	return "asynq"
}

// Header 返回头部信息
func (s *Server) Header() transport.Header {
	return nil
}

func (s *Server) RequestHeader() transport.Header {
	return nil
}

func (s *Server) ReplyHeader() transport.Header {
	return nil
}
