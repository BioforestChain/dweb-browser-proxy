// ================================================================================
// Code generated by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"frpConfManagement/internal/model"
)

type (
	IUser interface {
		Create(ctx context.Context, in model.UserCreateInput) (err error)
		GenerateMD5ByIdentification(identification string) (string, error)
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
