package ping

import (
	"context"
	v1 "github.com/BioforestChain/dweb-browser-proxy/api/ping/v1"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/middleware"
	"github.com/gogf/gf/v2/frame/g"
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

func (c *Controller) Ping(ctx context.Context, req *v1.Req) (res *v1.Res, err error) {
	g.RequestFromCtx(ctx).Response.Write(middleware.Response{0, "Network is success!", nil})
	return
}
