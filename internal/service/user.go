// ================================================================================
// Code generated by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"database/sql"

	//"github.com/gogf/gf/v2/frame/g"

	//"github.com/gogf/gf/v2/frame/g"
	v1 "proxyServer/api/client/v1"
	"proxyServer/internal/model"
	"proxyServer/internal/model/do"

	"github.com/gogf/gf/v2/database/gdb"
)

type (
	IUser interface {
		IsDomainExist(ctx context.Context, in model.CheckUrlInput) bool
		IsUserExist(ctx context.Context, in model.CheckUserInput) bool
		Create(ctx context.Context, in model.UserCreateInput) (entity *v1.ClientRegRes, err error)
		InsertDevice(ctx context.Context, tx gdb.TX, reqData model.DataToDevice) (result sql.Result, err error)
		GetUserList(ctx context.Context, in model.UserQueryInput) (entities []*do.User, total int, err error)
		GetDomainInfo(ctx context.Context, in model.AppQueryInput) (entities *v1.ClientQueryRes, err error)
		GenerateMD5ByPublicKeyIdentification(identification string) (string, error)
		IsUserIdentificationAvailable(ctx context.Context, identification string) (bool, error)
		IsNameAvailable(ctx context.Context, Name string) (bool, error)
		GetUserId(ctx context.Context, Name string) (uint32, error)
		GetDeviceId(ctx context.Context, DeviceIdentification string) (int, error)
		IsDomainAvailable(ctx context.Context, domain string) (bool, error)
		CreateAppInfo(ctx context.Context, in model.UserAppInfoCreateInput) (err error)
	}
)

var (
	localUser IUser
)

func User() IUser {
	if localUser == nil {
		panic("implement not found for interface IUser, forgot register?")
	}
	return localUser
}

func RegisterUser(i IUser) {
	localUser = i
}
