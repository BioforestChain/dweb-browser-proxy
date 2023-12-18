package auth

import (
	"context"
	"fmt"
	v1 "github.com/BioforestChain/dweb-browser-proxy/api/client/v1"
	"github.com/BioforestChain/dweb-browser-proxy/internal/consts"
	timeHelper "github.com/BioforestChain/dweb-browser-proxy/internal/pkg/util/time"
	"github.com/BioforestChain/dweb-browser-proxy/internal/service"
	"github.com/gogf/gf/v2/frame/g"
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

// GenToken
//
//	@Description:
//	@receiver s
//	@param ctx
//	@param UserId
//	@param UserIdentification
//	@return res
func (s *sAuth) GenToken(ctx context.Context, UserId uint32, UserIdentification string) (res *v1.ClientUserTokenDataRes) {
	token, refreshToken, expireTime, _ := service.Middleware().GenToken(UserId, UserIdentification)
	res = new(v1.ClientUserTokenDataRes)
	//res.UserID = UserId
	res.UserIdentification = UserIdentification
	res.Token = token
	res.RefreshToken = refreshToken
	res.NowTime = timeHelper.Date(timeHelper.Time(), consts.DefaultDateFormat)
	res.ExpireTime = timeHelper.Date(expireTime, consts.DefaultDateFormat)
	return res
}

// RefreshToken
//
//	@Description:
//	@receiver s
//	@param ctx
//	@param req
//	@return res
func (s *sAuth) RefreshToken(ctx context.Context, req *v1.ClientRefreshTokenReq) (res *v1.ClientUserTokenDataRes) {
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("ClientRefreshTokenReq Validator", err)
	}
	//token, refreshToken, expireTime, userId, err := service.Middleware().RefreshToken(req.AccessToken, req.RefreshToken)
	refreshTokenRes, err := service.Middleware().RefreshToken(req.AccessToken, req.RefreshToken)
	if err != nil {
		fmt.Println("ClientRefreshToken", err)
	}
	res = new(v1.ClientUserTokenDataRes)
	//res.UserID = refreshTokenRes.UserID
	res.UserIdentification = refreshTokenRes.UserIdentification
	res.Token = refreshTokenRes.Token
	res.RefreshToken = refreshTokenRes.RefreshToken
	res.ExpireTime = timeHelper.Date(refreshTokenRes.ExpireTime, consts.DefaultDateFormat)
	res.NowTime = timeHelper.Date(timeHelper.Time(), consts.DefaultDateFormat)
	return res
}
