/*
*

	@author: bnqkl
	@since: 2023/6/8/008 9:37
	@desc: //TODO

*
*/
package model

import (
	"github.com/BioforestChain/dweb-browser-proxy/pkg/model"
)

// UserAppInfoCreateInput
// @Description: Domain,AppInfo

type UserQueryInput struct {
	model.PaginationSearch
	Id     uint32
	Domain string
	Name   string
}
type AppQueryInput struct {
	UserName             string
	AppName              string
	AppId                string
	DeviceIdentification string
}

type CheckUrlInput struct {
	Host string
}
type CheckDomainInput struct {
	Domain string
	UserId uint32
	AppId  string
}
type CheckDeviceInput struct {
	DeviceIdentification string
}
type CheckUserInput struct {
	UserIdentification string
}
