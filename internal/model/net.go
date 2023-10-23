/*
*

	@author: bnqkl
	@since: 2023/6/8/008 9:37
	@desc: //TODO

*
*/
package model

import "proxyServer/internal/dao"

type NetModuleCreateInputReq struct {
	NetId  string
	Domain string
}

type NetQueryInput struct {
	dao.PaginationSearch
	Id     uint32
	Domain string
	Name   string
}
type DomainReq struct {
	Domain string
}

type UserAppInfoCreateInput struct {
	UserId               uint32
	UserName             string
	AppName              string
	AppId                string
	DeviceIdentification string
	PublicKey            string
	IsInstall            uint32
	Domain               string
	Timestamp            string
	Remark               string
	Subdomain            string
}
