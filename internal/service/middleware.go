// ================================================================================
// Code generated by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"proxyServer/internal/model/entity"
)

type (
	IMiddleware interface {
		Auth(r *ghttp.Request)
		GenToken(userID uint32) (string, error)
		ParseToken(tokenString string) (*entity.MyClaims, error)
		RefreshToken(aToken, rToken string) (newAToken, newRToken string, err error)
		Response(r *ghttp.Request)
	}
)

var (
	localMiddleware IMiddleware
)

func Middleware() IMiddleware {
	if localMiddleware == nil {
		panic("implement not found for interface IMiddleware, forgot register?")
	}
	return localMiddleware
}

func RegisterMiddleware(i IMiddleware) {
	localMiddleware = i
}
