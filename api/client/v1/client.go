package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"proxyServer/internal/dao"
	"proxyServer/internal/model/do"
)

// 1. 用户注册
// 1.1 设备表，一个用户可以有多个设备(暂定)
type ClientRegReq struct {
	g.Meta               `path:"/user/client-reg" tags:"ClientRegService" method:"post" summary:"Sign up a new client"`
	Name                 string `v:"required"` //用户名，昵称？
	PublicKey            string `v:"required"` //公钥
	DeviceIdentification string `v:""`         //暂定入参imei码，待生成设备标识
	Remark               string
}

// 1.2  域名注册
type ClientDomainRegReq struct {
	g.Meta               `path:"/user/client-domain-reg" tags:"ClientDomainRegService" method:"post" summary:"A new client with domain"`
	UserName             string `v:"required"`        //用户名称
	AppName              string `v:"required"`        //app名称
	AppIdentification    string `v:"required"`        //app唯一标识
	DeviceIdentification string `v:""`                //设备唯一标识
	PublicKey            string `v:"required"`        //公钥
	Domain               string `v:"required|domain"` //域名
	//Identification string `v:"required"`        // 通过用户名，查到user_id，数据插入到app表里
	Remark string
}

type ClientRegRes struct {
	DeviceIdentification string `json:"device_identification"` //设备标识，也就是clientID
	UserId               uint32 `json:"user_id"`               //用户id
}

type ClientQueryReq struct {
	g.Meta               `path:"/user/client-query" tags:"ClientQueryService" method:"get" summary:"Query client"`
	UserName             string `v:"required"` //用户名称
	AppName              string `v:"required"` //app名称
	AppIdentification    string `v:"required"` //app唯一标识
	DeviceIdentification string
}

type ClientListQueryReq struct {
	g.Meta         `path:"/user/client-list" tags:"ClientListQueryService" method:"get" summary:"Query client list"`
	Domain         string
	Identification string
	dao.PaginationSearch
}

type ClientQueryRes struct {
	Domain         string `json:"domain"`             // 域名
	Identification string `json:"app_identification"` // app唯一标识
}
type ClientQueryListRes struct {
	List     []*do.User `json:"list"`      // 列表
	Page     int        `json:"page"`      // 分页码
	Total    int        `json:"total"`     // 数据总数
	LastPage int        `json:"last_page"` // 最后一页
}

type ClientRefreshTokenReq struct {
	g.Meta       `path:"/user/client-refresh-token" tags:"ClientRefreshTokenService" method:"get" summary:"Get client refresh token"`
	AccessToken  string `v:"required",json:"token"`
	RefreshToken string `json:"refresh_token"`
}
type ClientUserTokenDataRes struct {
	UserID       uint32 `json:"user_id"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	NowTime      string `json:"now_time"`
	ExpireTime   string `json:"expire_time"`
}
