package v1

import (
	"github.com/BioforestChain/dweb-browser-proxy/app/pubsub/model/entity"
	"github.com/gogf/gf/v2/frame/g"
)

type Req struct {
	g.Meta `path:"/ping" tags:"Ping" method:"get" summary:"You first ping api"`
}
type RegReq struct {
	g.Meta `path:"/pubsub/permission/reg" tags:"PermissionReg" method:"post" summary:"You first permission reg api"`
	Id     int64  `v:""`         //
	Topic  string `v:"required"` // 订阅的主题名
	Type   uint32 `v:"required"` // 权限类型: 0:无认证，1:acl，2:基于密码，3:基于角色，4:etc
	//Publisher      string `v:"required"` // 创建者，（若chat场景是群主）
	NetDomainNames string `v:"required"` // acl中被授权的netDomain
}
type Res struct {
	g.Meta `mime:"text/html" example:"string"`
}

type PubsubPermissionDetailRes struct {
	Id        int                     `json:"id"        description:""`
	Topic     string                  `json:"name"      description:"订阅的主题名"`
	Type      uint32                  `json:"type"      description:"权限类型: 0:无认证，1:acl，2:基于密码，3:基于角色，4:etc"`
	Publisher string                  `json:"publisher" description:"创建者"`
	List      []*entity.PubsubUserAcl `json:"list"`
	//CreatedAt *gtime.Time             `json:"createdAt" description:"Created Time"`
	//UpdateAt  *gtime.Time             `json:"updateAt"  description:"Updated Time"`
}
