package middlewares

import "github.com/go-kratos/kratos/v2/middleware"

type Builder interface {
	Build() middleware.Middleware
}

type HttpBuilder struct {
	builders []Builder
}

func NewHttpBuilder(userAuth *UserAuth) *HttpBuilder {
	return &HttpBuilder{
		builders: []Builder{
			userAuth,
		},
	}
}
func (builder *HttpBuilder) Build() []middleware.Middleware {
	var middlewares []middleware.Middleware
	for _, builder := range builder.builders {
		middlewares = append(middlewares, builder.Build())
	}
	return middlewares
}
