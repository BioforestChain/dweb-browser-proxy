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

// ClientReg is the API for client sign up.
//
//	@Description:
//	@receiver c
//	@param ctx
//	@param req
//	@return res
//	@return err
func (c *Controller) ClientReg(ctx context.Context, req *v1.ClientRegReq) (res *v1.ClientRegRes, err error) {
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("clientReg Validator", err)
	}
	out, err := service.User().Create(ctx, model.UserCreateInput{
		Name:           req.Name,
		PublicKey:      req.PublicKey,
		Identification: req.DeviceIdentification,
		Remark:         req.Remark,
	})
	if err != nil {
		return
	}
	return &v1.ClientRegRes{
		out.DeviceIdentification,
	}, err
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
	res = &v1.ClientQueryListRes{
		Total:    total,
		Page:     condition.Page,
		List:     list,
		LastPage: commonLogic.GetLastPage(int64(total), condition.Limit),
	}
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
	return &v1.ClientQueryRes{
		data.Domain,
		data.Identification,
	}, err
}
