package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"proxyServer/internal/dao"
	"proxyServer/internal/model/do"
)

// ClientRegReq
// @Description:
// 1. 用户注册
// 1.1 设备表，一个用户可以有多个设备(暂定)
type ClientRegReq struct {
	g.Meta  `path:"/pre-user/client-reg" tags:"ClientRegService" method:"post" summary:"Sign up a new client"`
	Name    string ``             //用户名，昵称可填可不填
	UserKey string `v:"required"` //dweb中key模块申请的类uuid的字符串
	Remark  string
}

// ClientDomainReg
type ClientDomainRegReq struct {
	g.Meta            `path:"/user/client-domain-reg" tags:"ClientDomainRegService" method:"post" summary:"A new client domain registration"`
	UserName          string `v:""`         //用户名称
	AppName           string `v:"required"` //app名称
	AppIdentification string `v:"required"` //app唯一标识
	Subdomain         string `v:""`         //域名 todo 系统分配
	PublicKey         string `v:"required"` //公钥
	TTL               uint32 `v:""`         //存活时间
	Remark            string
}

// ClientAppInfoReportReq
// @Description:
// 1.2  App
type ClientAppInfoReportReq struct {
	g.Meta               `path:"/user/client-app-report" tags:"ClientAppInfoReportService" method:"post" summary:"A new client App report"`
	UserName             string `v:"required"` //用户名称
	AppName              string `v:"required"` //app名称
	AppIdentification    string `v:"required"` //app唯一标识
	DeviceIdentification string `v:""`         //设备唯一标识
	PublicKey            string `v:"required"` //公钥
	IsInstall            uint32 `v:"required"` //安装状态：1安装，0未安装
	Remark               string
}

// ClientRegRes
// @Description:
type ClientRegRes struct {
	UserIdentification string `json:"user_identification"` //用户id
	UserId             uint32 `json:"user_id"`             //db中用户id
}

// ClientQueryReq
// @Description:
type ClientQueryReq struct {
	g.Meta               `path:"/user/client-query" tags:"ClientQueryService" method:"get" summary:"Query client"`
	UserName             string `v:"required"` //用户名称
	AppName              string `v:"required"` //app名称
	AppIdentification    string `v:"required"` //app唯一标识
	DeviceIdentification string
}

// ClientListQueryReq
// @Description:
type ClientListQueryReq struct {
	g.Meta         `path:"/user/client-list" tags:"ClientListQueryService" method:"get" summary:"Query client list"`
	Domain         string
	Identification string
	dao.PaginationSearch
}

// ClientQueryRes
// @Description:
type ClientQueryRes struct {
	Domain         string `json:"domain"`             // 域名
	Identification string `json:"app_identification"` // app唯一标识
}

// ClientDomainQueryRes
// @Description:
type ClientDomainQueryRes struct {
	Domain string `json:"domain"` // 域名
}

// ClientQueryListRes
// @Description:
type ClientQueryListRes struct {
	List     []*do.User `json:"list"`      // 列表
	Page     int        `json:"page"`      // 分页码
	Total    int        `json:"total"`     // 数据总数
	LastPage int        `json:"last_page"` // 最后一页
}

// ClientRefreshTokenReq
// @Description:
type ClientRefreshTokenReq struct {
	g.Meta       `path:"/user/client-refresh-token" tags:"ClientRefreshTokenService" method:"get" summary:"Get client refresh token"`
	AccessToken  string `v:"required" ,json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// ClientUserTokenDataRes
// @Description:
type ClientUserTokenDataRes struct {
	UserIdentification string `json:"user_identification"`
	Token              string `json:"token"`
	RefreshToken       string `json:"refresh_token"`
	NowTime            string `json:"now_time"`
	ExpireTime         string `json:"expire_time"`
}

// ClientUserRefreshTokenRes
// @Description:
type ClientUserRefreshTokenRes struct {
	UserID             uint32 `json:"user_id"`
	UserIdentification string `json:"user_identification"`
	Token              string `json:"token"`
	RefreshToken       string `json:"refresh_token"`
	ExpireTime         int64  `json:"expire_time"`
}
