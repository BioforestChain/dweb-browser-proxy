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
		fmt.Println("req Validator", err)
	}

	err = service.User().Create(ctx, model.UserCreateInput{
		Name:           req.Name,
		Domain:         req.Domain,
		PublicKey:      req.PublicKey,
		Identification: req.Identification,
		Remark:         req.Remark,
	})
	return
}

func (c *Controller) ClientQuery(ctx context.Context, req *v1.ClientQueryReq) (res *v1.ClientQueryListRes, err error) {
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
