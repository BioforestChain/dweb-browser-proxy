package ping

import (
	"context"
	"proxyServer/internal/logic/middleware"

	"github.com/gogf/gf/v2/frame/g"

	v1 "proxyServer/api/ping/v1"
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

func (c *Controller) Ping(ctx context.Context, req *v1.Req) (res *v1.Res, err error) {
	g.RequestFromCtx(ctx).Response.Write(middleware.Response{0, "Network is success!", nil})
	return
}
