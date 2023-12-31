// =================================================================================
// Code generated by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// User is the golang structure for table user.
type User struct {
	Id             int64       `json:"id"             description:""`
	Name           string      `json:"name"           description:"用户名"`
	Remark         string      `json:"remark"         description:"备注信息"`
	Timestamp      int         `json:"timestamp"      description:"时间戳"`
	CreatedAt      *gtime.Time `json:"createdAt"      description:"Created Time"`
	UpdateAt       *gtime.Time `json:"updateAt"       description:"Updated Time"`
	DeletedAt      *gtime.Time `json:"deletedAt"      description:"Deleted Time"`
	UserKey        string      `json:"userKey"        description:""`
	Identification string      `json:"identification" description:"用户id"`
}
