package auth

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	v1 "proxyServer/api/client/v1"
	"proxyServer/internal/consts"
	timeHelper "proxyServer/internal/helper/time"
	"proxyServer/internal/service"
)

type (
	sAuth struct{}
)

func init() {
	service.RegisterAuth(New())
}
func New() service.IAuth {
	return &sAuth{}
}

func (s *sAuth) GenToken(ctx context.Context, UserId uint32) (res *v1.ClientUserTokenDataRes) {
	var user consts.User
	userId, _ := g.Cfg().Get(ctx, "auth.userId")
	user.UserID = userId.Uint32()

	token, refreshToken, expireTime, _ := service.Middleware().GenToken(UserId)
	res = new(v1.ClientUserTokenDataRes)
	res.UserID = UserId
	res.Token = token
	res.RefreshToken = refreshToken
	res.NowTime = timeHelper.Date(timeHelper.Time(), consts.DefaultDateFormat)
	res.ExpireTime = timeHelper.Date(expireTime, consts.DefaultDateFormat)
	return res
}

func (s *sAuth) RefreshToken(ctx context.Context, req *v1.ClientRefreshTokenReq) (res *v1.ClientUserTokenDataRes) {
	//var user consts.User
	//userId, _ := g.Cfg().Get(ctx, "auth.userId")
	//user.UserID = userId.Uint32()

	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("ClientRefreshTokenReq Validator", err)
	}

	token, refreshToken, expireTime, userId, err := service.Middleware().RefreshToken(req.AccessToken, req.RefreshToken)
	if err != nil {
		fmt.Println("ClientRefreshToken", err)
	}
	res = new(v1.ClientUserTokenDataRes)
	res.UserID = userId
	res.Token = token
	res.RefreshToken = refreshToken
	res.NowTime = timeHelper.Date(timeHelper.Time(), consts.DefaultDateFormat)
	res.ExpireTime = timeHelper.Date(expireTime, consts.DefaultDateFormat)
	return res
}
