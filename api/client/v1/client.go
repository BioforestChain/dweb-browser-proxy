package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"proxyServer/internal/dao"
	"proxyServer/internal/model/do"
)

type ClientRegReq struct {
	g.Meta         `path:"/user/client-reg" tags:"ClientRegService" method:"post" summary:"Sign up a new client"`
	Name           string `v:"required"`        //用户名，昵称？
	Domain         string `v:"required|domain"` // 域名
	PublicKey      string `v:"required"`
	Identification string `v:"required"` // 身份标识
	Remark         string
}
type ClientRegRes struct {
}

type ClientQueryReq struct {
	g.Meta         `path:"/user/client-query" tags:"ClientQueryService" method:"get" summary:"Query client"`
	Domain         string
	Identification string
	dao.PaginationSearch
}

type ClientQueryRes struct {
	List []*do.ProxyServerUser `json:"list"` // 列表
}
type ClientQueryListRes struct {
	List []*do.ProxyServerUser `json:"list"` // 列表
	//Stats map[string]int        `json:"stats"` // 搜索统计
	Page int `json:"page"` // 分页码
	//Size     int `json:"size"`     // 分页数量
	Total    int `json:"total"`    // 数据总数
	LastPage int `json:"lastPage"` // 最后一页
}

//type ContentSearchOutput struct {
//	List  []ContentSearchOutputItem `json:"list"`  // 列表
//	Stats map[string]int            `json:"stats"` // 搜索统计
//	Page  int                       `json:"page"`  // 分页码
//	Size  int                       `json:"size"`  // 分页数量
//	Total int                       `json:"total"` // 数据总数
//}
