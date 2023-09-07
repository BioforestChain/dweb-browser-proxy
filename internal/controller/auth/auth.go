package auth

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	v1 "proxyServer/api/client/v1"
	"proxyServer/internal/service"
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

// ClientGenToken is the API for client getToken.
//
//	@Description:
//	@receiver c
//	@param ctx
//	@param req
//	@return res
//	@return err
func (c *Controller) ClientGenToken(ctx context.Context, req *v1.ClientGenTokenReq) (res *v1.ClientUserTokenDataRes, err error) {
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("clientReg Validator", err)
	}
	out := service.Auth().GenToken(ctx)

	return &v1.ClientUserTokenDataRes{
		out.UserID,
		out.Token,
		out.NowTime,
	}, err
}
