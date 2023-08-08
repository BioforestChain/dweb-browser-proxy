// =================================================================================
// Code generated by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Device is the golang structure for table device.
type Device struct {
	Id             int         `json:"id"             description:""`
	UserId         int64       `json:"userId"         description:"用户id"`
	Name           string      `json:"name"           description:"名称"`
	Identification string      `json:"identification" description:"设备标识"`
	Remark         string      `json:"remark"         description:"备注信息"`
	Timestamp      string      `json:"timestamp"      description:"时间戳"`
	CreatedAt      *gtime.Time `json:"createdAt"      description:"Created Time"`
	UpdateAt       *gtime.Time `json:"updateAt"       description:"Updated Time"`
	DeletedAt      *gtime.Time `json:"deletedAt"      description:"Deleted Time"`
}