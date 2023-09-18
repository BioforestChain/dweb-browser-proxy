/*
*

	@author: bnqkl
	@since: 2023/6/8/008 9:37
	@desc: //TODO

*
*/
package model

import "proxyServer/internal/dao"

type UserCreateInput struct {
	Name           string
	PublicKey      string
	Identification string
	Domain         string
	Timestamp      string
	Remark         string
}

type UserAppInfoCreateInput struct {
	UserId               uint32
	UserName             string
	AppName              string
	AppIdentification    string
	DeviceIdentification string
	PublicKey            string
	IsInstall            uint32
	//Domain               string
	Timestamp string
	Remark    string
}

type UserQueryInput struct {
	dao.PaginationSearch
	Id     uint32
	Domain string
	Name   string
}
type AppQueryInput struct {
	UserName             string
	AppName              string
	AppIdentification    string
	DeviceIdentification string
}

type CheckUrlInput struct {
	Host string
}
type CheckDeviceInput struct {
	DeviceIdentification string
}
type CheckUserInput struct {
	UserIdentification string
}

type DataToDevice struct {
	UserId uint32 // 用户id
	//Name           interface{} // 名称
	SrcIdentification interface{} // 源设备标识
	Identification    interface{} // md5后设备标识
	Remark            interface{} // 备注信息
	Timestamp         interface{} // 时间戳
}
