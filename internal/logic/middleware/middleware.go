package middleware

import (
	"context"
	"proxyServer/internal/service"
)

type sMiddleware struct {
	Ctx    context.Context
	GCtx   *context.Context
	userId uint32
}

func init() {
	service.RegisterMiddleware(New())
}
func New() *sMiddleware {
	return &sMiddleware{}
}
