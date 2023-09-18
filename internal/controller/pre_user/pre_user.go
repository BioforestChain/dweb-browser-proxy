package pre_user

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

func (c *Controller) ClientReg(ctx context.Context, req *v1.ClientRegReq) (res *v1.ClientUserTokenDataRes, err error) {
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("clientReg Validator", err)
	}
	newOne, err := service.User().Create(ctx, model.UserCreateInput{
		Name:      req.Name,
		PublicKey: req.PublicKey,
		//Identification: req.DeviceIdentification,
		Remark: req.Remark,
	})
	if err != nil {
		return
	}
	out := service.Auth().GenToken(ctx, newOne.UserId, newOne.UserIdentification)
	return &v1.ClientUserTokenDataRes{
		//newOne.UserId,
		newOne.UserIdentification,
		out.Token,
		out.RefreshToken,
		out.NowTime,
		out.ExpireTime,
	}, err
}
