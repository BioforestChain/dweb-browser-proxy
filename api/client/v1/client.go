package v1

import (
	"github.com/BioforestChain/dweb-browser-proxy/internal/dao"
	"github.com/BioforestChain/dweb-browser-proxy/internal/model/do"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

//// ClientRegReq
//// @Description:
//// 1. 用户注册
//// 1.1 用户表，一个用户可以有多个设备(暂定)
//type ClientRegReq struct {
//	g.Meta  `path:"/pre-user/client-reg" tags:"ClientRegService" method:"post" summary:"Sign up a new client"`
//	Name    string ``             //用户名，昵称可填可不填
//	UserKey string `v:"required"` //dweb中key模块申请的类uuid的字符串
//	Remark  string
//}

type ClientNetModuleRegReq struct {
	g.Meta           `path:"/user/net-module-reg" tags:"ClientNetModuleRegService" method:"post" summary:"Sign up a new client net-module"`
	Id               int64  `v:""`         //
	ServerAddr       string `v:"required"` //服务地址
	Port             uint32 `v:"required"` //端口
	Secret           string `v:"required"` //密钥
	BroadcastAddress string `v:"required"` //广播地址
	//RootDomain       string `v:"required"` //
	NetId string `v:"required"` //网络模块id
	U     string `v:"required"` //设备的 uuid
}
type ClientNetModuleDetailReq struct {
	g.Meta `path:"/user/net-module-detail" tags:"ClientNetModuleDetailService" method:"get" summary:"Get net-module detail by id"`
	Id     uint32 `v:"bail|required|integer"` //id
}

type ClientNetModuleDetailRes struct {
	Id                     int64       `json:"id"`                       //db 主键id
	NetId                  string      `json:"net_id"`                   //网络模块id
	Domain                 string      `json:"domain"`                   // 域名
	RootDomain             string      `json:"root_domain"`              // 根域名
	PrefixBroadcastAddress string      `json:"prefix_broadcast_address"` // 广播地址前缀
	BroadcastAddress       string      `json:"broadcast_address"`        // 广播地址
	Port                   interface{} `json:"port"`                     // 端口
	Remark                 interface{} `json:"remark,omitempty"`         // 备注信息
	Timestamp              interface{} `json:"timestamp"`                // 时间戳
	IsOnline               interface{} `json:"is_online"`                // 上线：1上线，0下线
	CreatedAt              *gtime.Time `json:"created_at"`               // Created Time
	UpdateAt               *gtime.Time `json:"update_at"`                // Updated Time
	IsSelected             interface{} `json:"is_selected"`              // 是否选中
	PrivateKey             interface{} `json:"private_key"`              // rsa私钥
	PublicKey              interface{} `json:"public_key"`               // rsa公钥
}

type ClientNetModuleListReq struct {
	g.Meta   `path:"/user/net-module-list" tags:"ClientNetModuleListService" method:"get" summary:"Query net-module list"`
	Domain   string `json:"domain"`
	NetId    string `json:"net_id"`
	IsOnline uint32 `json:""`
	dao.PaginationSearch
}

type ClientNetModuleListRes struct {
	List     []*ClientNetModuleDetailRes `json:"list"`      // 列表
	Page     int                         `json:"page"`      // 分页码
	Total    int                         `json:"total"`     // 数据总数
	LastPage int                         `json:"last_page"` // 最后一页
}

type ClientAppModuleRegReq struct {
	g.Meta         `path:"/user/app-module-reg" tags:"ClientAppModuleRegService" method:"post" summary:"Sign up a new client app-module with the net-module"`
	ArrayAppIdInfo string `v:"required"` //app模块id
	//AppId    string `v:"required"` //app模块id
	//NetId    string `v:"required"` //网络模块id
	//UserName string `v:"required"` //用户名
	//AppName  string `v:"required"` //App模块名
}

type ClientAppModuleRegRes struct {
	Id        int64       `json:"id"`          //db 主键id
	NetId     string      `json:"net_id"`      //网络模块id
	AppId     string      `json:"app_id"`      //App模块id
	UserName  interface{} `json:"user_name"`   // 用户名称
	AppName   interface{} `json:"app_name"`    // 模块名称
	Timestamp interface{} `json:"timestamp"`   // 时间戳
	CreatedAt *gtime.Time `json:"created_at"`  // Created Time
	UpdateAt  *gtime.Time `json:"update_at"`   // Updated Time
	CumReqNum interface{} `json:"cum_req_num"` // 累计被请求次数
	IsInstall interface{} `json:"is_install"`  // 是否安装：1安装，0（未安装）卸载
	IsOnline  interface{} `json:"is_online"`   // 是否在线：1在线，0不在线
	PublicKey interface{} `json:"public_key"`  // 公钥
}

// ClientAppInfoReportReq
// @Description:
// 1.2  App
type ClientAppInfoReportReq struct {
	g.Meta    `path:"/user/client-app-report" tags:"ClientAppInfoReportService" method:"post" summary:"A new client App report"`
	UserName  string `v:"required"` //用户名称
	AppName   string `v:""`         //App模块名称
	AppId     string `v:"required"` //App模块id
	NetId     string `v:"required"` //网络模块id
	PublicKey string `v:""`         //公钥
	IsInstall uint32 `v:"required"` //安装状态：1 安装，0 未安装
	IsOnline  uint32 `v:"required"` //上线状态：1 上线，0 下线
	Remark    string
}

type ClientAppInfoReportRes struct {
}

type ClientAppModuleDelReq struct {
	g.Meta `path:"/user/app-module-del" tags:"ClientAppModuleDelService" method:"post" summary:"Del app-module by id"`
	Id     int64 `v:"required"`
}

type ClientAppModuleListReq struct {
	g.Meta    `path:"/user/app-module-list" tags:"ClientAppModuleListService" method:"get" summary:"Query app-module list"`
	UserName  string `json:"user_name"   in:"query" dc:"用户名"`
	NetId     string `json:"net_id"   in:"query" dc:"网络模块id"`
	AppId     string `json:"app_id"   in:"query" dc:"app模块id"`
	AppName   string `json:"app_name"   in:"query" dc:"app模块名"`
	IsInstall uint32 `json:"is_install"   in:"query" dc:"是否安装"`
	IsOnline  uint32 `json:"is_online"   in:"query" dc:"是否在线"`
	dao.PaginationSearch
}

type ClientAppModuleDetailRes struct {
	Id        int64       `json:"id"`          //db 主键id
	AppId     string      `json:"app_id"`      //App模块id
	NetId     string      `json:"net_id"`      //网络模块id
	UserName  string      `json:"user_name"`   // 域名
	AppName   string      `json:"app_name"`    // 域名
	Timestamp interface{} `json:"timestamp"`   // 时间戳
	CreatedAt *gtime.Time `json:"created_at"`  // Created Time
	UpdateAt  *gtime.Time `json:"update_at"`   // Updated Time
	CumReqNum interface{} `json:"cum_req_num"` // 累计被请求次数
	IsInstall interface{} `json:"is_install"`  // 是否安装：1安装，0（未安装）卸载
	IsOnline  interface{} `json:"is_online"`   // 是否在线：1在线，0不在线
	PublicKey interface{} `json:"public_key"`  // 公钥
}
type ClientAppModuleListRes struct {
	List     []*ClientAppModuleDetailRes `json:"list"`      // 列表
	Page     int                         `json:"page"`      // 分页码
	Total    int                         `json:"total"`     // 数据总数
	LastPage int                         `json:"last_page"` // 最后一页
}

// ClientQueryReq
// @Description:
type ClientQueryReq struct {
	g.Meta               `path:"/user/client-query" tags:"ClientQueryService" method:"get" summary:"Query client"`
	UserName             string `v:"required"` //用户名称
	AppName              string `v:"required"` //App模块名称
	AppIdentification    string `v:"required"` //App模块id
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
	Domain *gvar.Var `json:"domain"` // 域名
	//Identification string        `json:"app_identification"` // App模块id
}

// ClientDomainQueryRes
// @Description:
type ClientDomainQueryRes struct {
	Domain string `json:"domain"` // 域名
}

// ClientQueryListRes
// @Description:
type ClientQueryListRes struct {
	List     []*do.Net `json:"list"`      // 列表
	Page     int       `json:"page"`      // 分页码
	Total    int       `json:"total"`     // 数据总数
	LastPage int       `json:"last_page"` // 最后一页
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
