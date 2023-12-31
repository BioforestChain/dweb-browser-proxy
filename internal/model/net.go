/*
*

	@author: bnqkl
	@since: 2023/6/8/008 9:37
	@desc: //TODO

*
*/
package model

import (
	"proxyServer/internal/dao"
)

type NetModuleCreateInput struct {
	Id         int64
	NetId      string
	Domain     string
	RootDomain string
	Secret     string
	Port       uint32
}
type NetModuleDetailInput struct {
	Id uint32
}
type NetModuleListQueryInput struct {
	dao.PaginationSearch
	Domain   string `json:"domain"   in:"query" dc:"域名"`
	NetId    string `json:"net_id"   in:"query" dc:"网络模块id"`
	IsOnline uint32 `json:"is_online"   in:"query" dc:"是否上线"`
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
