package pubsub_permission

import (
	"context"
	"fmt"
	v1 "github.com/BioforestChain/dweb-browser-proxy/app/pubsub/api/pubsub_permission/v1"
	"github.com/BioforestChain/dweb-browser-proxy/app/pubsub/consts"
	"github.com/BioforestChain/dweb-browser-proxy/app/pubsub/model"
	"github.com/BioforestChain/dweb-browser-proxy/app/pubsub/service"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/middleware"
	"github.com/gogf/gf/v2/frame/g"
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

func (c *Controller) Ping(ctx context.Context, req *v1.Req) (res *v1.Res, err error) {
	g.RequestFromCtx(ctx).Response.Write(middleware.Response{0, "permission is success!", nil})
	return
}

// Reg
//
//	@Description: 权限注册
//	@receiver c
//	@param ctx
//	@param req
//	@return res
//	@return err
func (c *Controller) Reg(ctx context.Context, req *v1.RegReq) (res *v1.PubsubPermissionDetailRes, err error) {

	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("NetModuleReg Validator", err)
		return nil, err
	}

	res, err = service.PubsubPermission().CreatePubsubPermission(ctx, model.PubsubPermissionCreateInput{
		Id:             req.Id,
		Topic:          req.Topic,
		Type:           req.Type,
		NetDomainNames: req.NetDomainNames,
		XDwebHostMMID:  g.RequestFromCtx(ctx).Header.Get(consts.XDwebHostMMID),
	})
	return
}
