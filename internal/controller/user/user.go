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

// SignUp is the API for user sign up.
func (c *Controller) ClientReg(ctx context.Context, req *v1.ClientRegReq) (res *v1.ClientRegRes, err error) {
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("clientReg Validator", err)
	}

	err = service.User().Create(ctx, model.UserCreateInput{
		Name:           req.Name,
		PublicKey:      req.PublicKey,
		Identification: req.DeviceIdentification,
		Remark:         req.Remark,
	})
	return
}

// 1.2 域名注册
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

func (c *Controller) ClientQuery(ctx context.Context, req *v1.ClientQueryReq) (res *v1.ClientQueryRes, err error) {
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("ClientQuery Validator", err)
	}
	//condition := model.AppQueryInput{}
	//UserName             string
	//AppName              string
	//AppIdentification    string
	//DeviceIdentification string

	data, err := service.User().GetDomainInfo(ctx, model.AppQueryInput{
		AppIdentification:    req.AppIdentification,
		DeviceIdentification: req.DeviceIdentification,
		UserName:             req.UserName,
		AppName:              req.AppName,
	})
	res = &v1.ClientQueryRes{
		data.Domain,
		data.Identification,
	}
	return res, err
}
