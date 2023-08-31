package auth

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	v1 "proxyServer/api/client/v1"
	"proxyServer/internal/consts"
	"proxyServer/internal/service"

	"time"
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

func (s *sAuth) GenToken(ctx context.Context) (res *v1.ClientUserTokenDataRes) {
	var user consts.User
	userId, _ := g.Cfg().Get(ctx, "auth.userId")
	user.UserID = userId.Uint32()
	token, _ := service.Middleware().GenToken(user.UserID)
	res = new(v1.ClientUserTokenDataRes)
	res.UserID = user.UserID
	res.Token = token
	res.NowTime = int(time.Now().UnixMilli())
	return res
}
