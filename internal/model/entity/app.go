// =================================================================================
// Code generated by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// App is the golang structure for table app.
type App struct {
	Id             int         `json:"id"             description:""`
	UserId         int64       `json:"userId"         description:"用户id"`
	DeviceId       int64       `json:"deviceId"       description:"设备id"`
	Name           string      `json:"name"           description:"名称"`
	Identification string      `json:"identification" description:"app唯一标识"`
	Remark         string      `json:"remark"         description:"备注信息"`
	Timestamp      string      `json:"timestamp"      description:"时间戳"`
	CreatedAt      *gtime.Time `json:"createdAt"      description:"Created Time"`
	UpdateAt       *gtime.Time `json:"updateAt"       description:"Updated Time"`
	DeletedAt      *gtime.Time `json:"deletedAt"      description:"Deleted Time"`
	CumReqNum      int         `json:"cumReqNum"      description:"累计被请求次数"`
	IsInstall      int         `json:"isInstall"      description:"是否安装：1安装0（未安装）卸载"`
}
