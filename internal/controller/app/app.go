package app

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

// AppModuleReg
//
//	@Description:
//	@receiver c
//	@param ctx
//	@param req
//	@return res
//	@return err

func (c *Controller) AppModuleReg(ctx context.Context, req *v1.ClientAppModuleRegReq) (res *v1.ClientAppModuleRegRes, err error) {
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("AppModuleReg Validator", err)
	}
	res, err = service.App().CreateAppModule(ctx, model.AppModuleCreateInput{
		NetId:    req.NetId,
		AppId:    req.AppId,
		UserName: req.UserName,
		AppName:  req.AppName,
	})
	return
}

// ClientAppInfoReport
//
//	@Description: App信息上报
//	@receiver c
//	@param ctx
//	@param req
//	@return res
//	@return err
func (c *Controller) ClientAppInfoReport(ctx context.Context, req *v1.ClientAppInfoReportReq) (res *v1.ClientAppInfoReportRes, err error) {
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("ClientAppInfoReport Validator", err)
	}
	err = service.App().CreateAppInfo(ctx, model.AppModuleInfoCreateInput{
		UserName:  req.UserName,
		AppId:     req.AppId,
		NetId:     req.NetId,
		PublicKey: req.PublicKey,
		AppName:   req.AppName,
		IsInstall: req.IsInstall,
		IsOnline:  req.IsOnline,
		Remark:    req.Remark,
	})
	return
}
