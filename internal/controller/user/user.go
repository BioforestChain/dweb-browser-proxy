package user

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	v1 "proxyServer/api/client/v1"
	commonLogic "proxyServer/internal/logic"
	"proxyServer/internal/model"
	"proxyServer/internal/service"
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

// ClientDomainReg
//
//	@Description: 域名注册
//	@receiver c
//	@param ctx
//	@param req
//	@return res
//	@return err
func (c *Controller) ClientDomainReg(ctx context.Context, req *v1.ClientDomainRegReq) (res *v1.ClientRegRes, err error) {
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("ClientDomainReg Validator", err)
	}
	var getUserId uint32
	err = service.User().CreateDomain(ctx, model.UserDomainCreateInput{
		UserId:               getUserId,
		UserName:             req.UserName,
		AppIdentification:    req.AppIdentification,
		DeviceIdentification: req.DeviceIdentification,
		PublicKey:            req.PublicKey,
		AppName:              req.AppName,
		Domain:               req.Domain,
		Remark:               req.Remark,
	})
	return
}

// ClientListQuery
//
//	@Description: 获取用户列表
//	@receiver c
//	@param ctx
//	@param req
//	@return res
//	@return err
func (c *Controller) ClientListQuery(ctx context.Context, req *v1.ClientListQueryReq) (res *v1.ClientQueryListRes, err error) {
	condition := model.UserQueryInput{}
	condition.Page, condition.Limit, condition.Offset = commonLogic.InitCodion(condition.Page, condition.Limit)
	condition.Domain = req.Domain
	list, total, err := service.User().GetUserList(ctx, condition)

	res = new(v1.ClientQueryListRes)
	res.Total = total
	res.List = list
	res.Page = condition.Page
	res.LastPage = commonLogic.GetLastPage(int64(total), condition.Limit)
	return res, err
}

// ClientQuery
//
//	@Description: 查询用户应用的域名等信息
//	@receiver c
//	@param ctx
//	@param req
//	@return res
//	@return err
func (c *Controller) ClientQuery(ctx context.Context, req *v1.ClientQueryReq) (res *v1.ClientQueryRes, err error) {
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("ClientQuery Validator", err)
	}
	//condition := model.AppQueryInput{}
	//app,device标识,用户名,app名
	data, err := service.User().GetDomainInfo(ctx, model.AppQueryInput{
		AppIdentification:    req.AppIdentification,
		DeviceIdentification: req.DeviceIdentification,
		UserName:             req.UserName,
		AppName:              req.AppName,
	})
	if err != nil {
		return
	}
	if data != nil {
		res = new(v1.ClientQueryRes)
		res.Domain = data.Domain
		res.Identification = data.Identification
		return res, err
	}
	return
}
