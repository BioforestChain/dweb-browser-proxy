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
	Timestamp      string
	Remark         string
}

type UserDomainCreateInput struct {
	UserId               uint32
	UserName             string
	AppName              string
	AppIdentification    string
	DeviceIdentification string
	Domain               string
	Timestamp            string
	Remark               string
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
