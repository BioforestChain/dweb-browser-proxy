package app

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "github.com/BioforestChain/dweb-browser-proxy/api/client/v1"
	"github.com/BioforestChain/dweb-browser-proxy/internal/model"
	"github.com/BioforestChain/dweb-browser-proxy/internal/service"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/page"
	"github.com/gogf/gf/v2/frame/g"
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
	var appInfos []map[string]string
	json.Unmarshal([]byte(req.ArrayAppIdInfo), &appInfos)
	for _, v := range appInfos {
		res, err = service.App().CreateAppModule(ctx, model.AppModuleCreateInput{
			NetId:    v["netId"],
			AppId:    v["appId"],
			UserName: v["userName"],
			AppName:  v["appName"],
		})
	}
	return
}
func (c *Controller) AppModuleDel(ctx context.Context, req *v1.ClientAppModuleDelReq) (res *v1.ClientAppModuleRegRes, err error) {
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("AppModuleDel Validator", err)
	}
	err = service.App().DelAppById(ctx, model.AppModuleDelInput{
		Id: req.Id,
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

// AppModuleList
//
//	@Description:
//	@receiver c
//	@param ctx
//	@param req
//	@return res
//	@return err
func (c *Controller) AppModuleList(ctx context.Context, req *v1.ClientAppModuleListReq) (res *v1.ClientAppModuleListRes, err error) {
	condition := model.AppModuleListQueryInput{}
	condition.Page, condition.Limit, condition.Offset = page.InitCondition(req.Page, req.Limit)
	condition.UserName = req.UserName
	condition.NetId = req.NetId
	condition.AppId = req.AppId
	condition.AppName = req.AppName
	condition.IsInstall = req.IsInstall
	condition.IsOnline = req.IsOnline

	list, total, err := service.App().GetAppModuleList(ctx, condition)
	if err != nil {
		return
	}
	//if list == nil {
	//	//list = []*v1.ClientAppModuleDetailRes{}
	//	list = make([]*v1.ClientAppModuleDetailRes, 0)
	//}
	res = new(v1.ClientAppModuleListRes)
	res.List = list
	res.Total = total
	res.Page = condition.Page
	res.LastPage = page.GetLastPage(int64(total), condition.Limit)
	return res, err
}
