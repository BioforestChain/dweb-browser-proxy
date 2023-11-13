package model

import "proxyServer/internal/dao"

type AppModuleCreateInput struct {
	NetId    string
	AppId    string
	UserName string
	AppName  string
}
type AppModuleDelInput struct {
	Id int64
}
type AppModuleInfoCreateInput struct {
	UserName  string
	AppName   string
	AppId     string
	NetId     string
	PublicKey string
	IsInstall uint32
	IsOnline  uint32
	Timestamp string
	Remark    string
}

type AppModuleListQueryInput struct {
	dao.PaginationSearch
	UserName  string `json:"user_name"   in:"query" dc:"用户名"`
	NetId     string `json:"net_id"   in:"query" dc:"网络模块id"`
	AppId     string `json:"app_id"   in:"query" dc:"app模块id"`
	AppName   string `json:"app_name"   in:"query" dc:"app模块名"`
	IsInstall uint32 `json:"is_install"   in:"query" dc:"是否安装"`
	IsOnline  uint32 `json:"is_online"   in:"query" dc:"是否在线"`
}
