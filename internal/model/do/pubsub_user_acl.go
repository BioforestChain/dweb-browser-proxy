// =================================================================================
// Code generated by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// PubsubUserAcl is the golang structure of table pubsub_user_acl for DAO operations like Where/Data.
type PubsubUserAcl struct {
	g.Meta       `orm:"table:pubsub_user_acl, do:true"`
	Id           interface{} //
	PermissionId interface{} // permission主键
	NetDomain    interface{} // netDomain
	CreatedAt    *gtime.Time // Created Time
	UpdateAt     *gtime.Time // Updated Time
}
