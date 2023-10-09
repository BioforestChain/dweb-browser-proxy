package pre_user

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	v1 "proxyServer/api/client/v1"
	"proxyServer/internal/model"
	"proxyServer/internal/service"
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

func (c *Controller) ClientReg(ctx context.Context, req *v1.ClientRegReq) (res *v1.ClientUserTokenDataRes, err error) {
	var rule = `regex:^[a-zA-Z0-9]{3,32}$|max-length:32`
	if err := g.Validator().Rules(rule).Data(req.Name).Run(ctx); err != nil {
		fmt.Println("clientReg Name Validator", err.Error())
		return nil, gerror.Newf(`The value "%s" must be letters and digits complies with domain name rules`, req.Name)
	}
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("clientReg Validator", err)
	}
	newOne, err := service.User().CreateUser(ctx, model.UserCreateInput{
		Name:    req.Name,
		UserKey: req.UserKey,
		Remark:  req.Remark,
	})
	if err != nil {
		return
	}
	out := service.Auth().GenToken(ctx, newOne.UserId, newOne.UserIdentification)
	return &v1.ClientUserTokenDataRes{
		newOne.UserIdentification,
		out.Token,
		out.RefreshToken,
		out.NowTime,
		out.ExpireTime,
	}, err
}
