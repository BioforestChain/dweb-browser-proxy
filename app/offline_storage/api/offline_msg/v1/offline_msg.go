package v1

import (
	"github.com/BioforestChain/dweb-browser-proxy/pkg/model"
	"github.com/gogf/gf/v2/frame/g"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Req struct {
	g.Meta   `path:"/user/offline-msg-list" tags:"Ping" method:"get" summary:"offline-msg-list-api"`
	ClientID string `v:"required"` // 接收者ID,用户的域名，广播地址
	model.PaginationSearch
}
type Res struct {
	//Id any      `json:"_id"`     // db 主键id
	Id primitive.ObjectID `bson:"_id"`
	C  string             `json:"content"` // 离线消息内容
	RS []string           `json:"rs"`
	//RS string `json:"receiver"` // 接收者
	//Content interface{} `json:"content"   in:"query" dc:"离线消息内容"`
	D string `json:"date"   in:"query" dc:"时间"`
}

type OfflineMsgListRes struct {
	List     []*Res `json:"list"`      // 列表
	Page     int    `json:"page"`      // 分页码
	Total    int    `json:"total"`     // 数据总数
	LastPage int    `json:"last_page"` // 最后一页
}
