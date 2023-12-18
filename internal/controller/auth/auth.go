package auth

import (
	"context"
	"fmt"
	v1 "github.com/BioforestChain/dweb-browser-proxy/api/client/v1"
	"github.com/BioforestChain/dweb-browser-proxy/internal/service"
	"github.com/gogf/gf/v2/frame/g"
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
//func (c *Controller) ClientGenToken(ctx context.Context, req *v1.ClientGenTokenReq) (res *v1.ClientUserTokenDataRes, err error) {
//	if err := g.Validator().Data(req).Run(ctx); err != nil {
//		fmt.Println("ClientGenToken Validator", err)
//	}
//	out := service.Auth().GenToken(ctx)
//
//	return &v1.ClientUserTokenDataRes{
//		out.UserID,
//		out.Token,
//		out.RefreshToken,
//		out.NowTime,
//		out.ExpireTime,
//	}, err
//}

// ClientRefreshToken
//
//	@Description:
//	@receiver c
//	@param ctx
//	@param req
//	@return res
//	@return err
func (c *Controller) ClientRefreshToken(ctx context.Context, req *v1.ClientRefreshTokenReq) (res *v1.ClientUserTokenDataRes, err error) {
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("ClientRefreshTokenReq Validator", err)
	}
	out := service.Auth().RefreshToken(ctx, req)

	return &v1.ClientUserTokenDataRes{
		out.UserIdentification,
		out.Token,
		out.RefreshToken,
		out.NowTime,
		out.ExpireTime,
	}, err
}
