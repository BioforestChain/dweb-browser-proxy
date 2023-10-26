package net

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	v1 "proxyServer/api/client/v1"
	"proxyServer/internal/model"
	"proxyServer/internal/service"
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

// NetModuleReg
//
//	@Description:
//	@receiver c
//	@param ctx
//	@param req
//	@return res
//	@return err

func (c *Controller) NetModuleReg(ctx context.Context, req *v1.ClientNetModuleRegReq) (res *v1.ClientNetModuleRegRes, err error) {
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("NetModuleReg Validator", err)
	}
	res, err = service.Net().CreateNetModule(ctx, model.NetModuleCreateInputReq{
		NetId:  req.NetId,
		Domain: req.Domain,
	})
	return
}
