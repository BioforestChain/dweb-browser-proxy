package net

import (
	"context"
	"fmt"
	v1 "github.com/BioforestChain/dweb-browser-proxy/api/client/v1"
	"github.com/BioforestChain/dweb-browser-proxy/internal/model"
	"github.com/BioforestChain/dweb-browser-proxy/internal/pkg/page"
	"github.com/BioforestChain/dweb-browser-proxy/internal/service"
	"github.com/gogf/gf/v2/frame/g"
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

func (c *Controller) NetModuleReg(ctx context.Context, req *v1.ClientNetModuleRegReq) (res *v1.ClientNetModuleDetailRes, err error) {
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("NetModuleReg Validator", err)
		return nil, err
	}
	// 参数长度至多长度为9个，数字或者字母组合
	//pattern := regexp.MustCompile(`^[a-z\d]{1,9}$`)
	//matchDomain := pattern.MatchString(req.Domain)
	//if !matchDomain {
	//	return nil, gerror.Newf(`Sorry, your domain "%s" must be combination of lowercase letters and numbers,the length is 1 to 9 yet`, req.Domain)
	//}

	//rule := "regex:`[a-z\\d]{1,9}$`"
	//if err := g.Validator().Rules(rule).Data(req.Domain).Run(ctx); err != nil {
	//	fmt.Println("NetModuleReg Validator Domain", err)
	//	return nil, err
	//}

	res, err = service.Net().CreateNetModule(ctx, model.NetModuleCreateInput{
		Id:               req.Id,
		NetId:            req.NetId,
		ServerAddr:       req.ServerAddr,
		Secret:           req.Secret,
		Port:             req.Port,
		BroadcastAddress: req.BroadcastAddress,
	})
	return
}

func (c *Controller) NetModuleDetailById(ctx context.Context, req *v1.ClientNetModuleDetailReq) (res *v1.ClientNetModuleDetailRes, err error) {
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("NetModuleReg Validator", err)
		return nil, err
	}
	//pattern := regexp.MustCompile(`^[a-z\d]{1,9}$`)
	//matchDomain := pattern.MatchString(req.Domain)
	//if !matchDomain {
	//	return nil, gerror.Newf(`Sorry, your domain "%s" must be combination of lowercase letters and numbers,the length is 1 to 9 yet`, req.Domain)
	//}

	res, err = service.Net().GetNetModuleDetailById(ctx, model.NetModuleDetailInput{
		Id: req.Id,
	})
	return
}

// NetModuleList
//
//	@Description: 网络模块列表
//	@receiver c
//	@param ctx
//	@param req
//	@return res
//	@return err
func (c *Controller) NetModuleList(ctx context.Context, req *v1.ClientNetModuleListReq) (res *v1.ClientNetModuleListRes, err error) {
	condition := model.NetModuleListQueryInput{}
	condition.Page, condition.Limit, condition.Offset = page.InitCondition(req.Page, req.Limit)
	condition.NetId = req.NetId
	condition.Domain = req.Domain
	condition.IsOnline = req.IsOnline

	list, total, err := service.Net().GetNetModuleList(ctx, condition)
	res = new(v1.ClientNetModuleListRes)
	res.List = list
	res.Total = total
	res.Page = condition.Page
	res.LastPage = page.GetLastPage(int64(total), condition.Limit)
	return res, err
}

// NetModuleDel
//
//	@Description:
//	@receiver c
//	@param ctx
//	@param req
//	@return res
//	@return err
func (c *Controller) NetModuleDel(ctx context.Context, req *v1.ClientNetModuleListReq) (res *v1.ClientNetModuleListRes, err error) {
	condition := model.NetModuleListQueryInput{}
	condition.Page, condition.Limit, condition.Offset = page.InitCondition(req.Page, req.Limit)
	condition.NetId = req.NetId
	condition.Domain = req.Domain
	condition.IsOnline = req.IsOnline

	list, total, err := service.Net().GetNetModuleList(ctx, condition)
	res = new(v1.ClientNetModuleListRes)
	res.List = list
	res.Total = total
	res.Page = condition.Page
	res.LastPage = page.GetLastPage(int64(total), condition.Limit)
	return res, err
}
