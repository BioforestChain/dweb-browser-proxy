// =================================================================================
// Code generated by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// PubsubUserAcl is the golang structure for table pubsub_user_acl.
type PubsubUserAcl struct {
	Id           int         `json:"id"           description:""`
	PermissionId int         `json:"permissionId" description:"permission主键"`
	NetDomain    string      `json:"netDomain"    description:"netDomain"`
	CreatedAt    *gtime.Time `json:"createdAt"    description:"Created Time"`
	UpdateAt     *gtime.Time `json:"updateAt"     description:"Updated Time"`
}