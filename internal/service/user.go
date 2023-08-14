// ================================================================================
// Code generated by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	v1 "proxyServer/api/client/v1"
	"proxyServer/internal/model"
	"proxyServer/internal/model/do"
)

type (
	IUser interface {
		Create(ctx context.Context, in model.UserCreateInput) (err error)
		CreateDomain(ctx context.Context, in model.UserDomainCreateInput) (err error)
		GetUserList(ctx context.Context, in model.UserQueryInput) (out []*do.User,total int,err error)
		IsDomainExist(ctx context.Context, in model.CheckUrlInput) (bool)
		GetDomainInfo(ctx context.Context, in model.AppQueryInput) (out *v1.ClientQueryRes,err error)
		GenerateMD5ByDeviceIdentification(identification string) (string, error)
		IsIdentificationAvailable(ctx context.Context, identification string) (bool, error)
		IsNameAvailable(ctx context.Context, Name string) (bool, error)
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
