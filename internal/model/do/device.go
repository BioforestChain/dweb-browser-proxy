// =================================================================================
// Code generated by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Device is the golang structure of table device for DAO operations like Where/Data.
//type Device struct {
//	g.Meta         `orm:"table:device, do:true"`
//	Id             interface{} //
//	UserId         interface{} // 用户id
//	Name           interface{} // 名称
//	Identification interface{} `gorm:"column:identification" db:"identification" json:"identification" form:"identification"` // 设备标识
//	Remark         interface{} // 备注信息
//	Timestamp      interface{} // 时间戳
//	CreatedAt      *gtime.Time // Created Time
//	UpdateAt       *gtime.Time // Updated Time
//	DeletedAt      *gtime.Time // Deleted Time
//}

// Device is the golang structure of table device for DAO operations like Where/Data.
type Device struct {
	g.Meta         `orm:"table:device, do:true"`
	Id interface{} `json:"id"`
	UserId interface{} `json:"user_id"` // 用户id
	Name interface{} `json:"name,omitempty"` // 名称
	SrcIdentification interface{} `json:"src_identification,omitempty"` // 源设备标识
	Identification interface{} `json:"identification,omitempty"` // 设备标识
	Remark interface{} `json:"remark,omitempty"` // 备注信息
	Timestamp interface{} `json:"timestamp"` // 时间戳
	CreatedAt *gtime.Time `json:"created_at,omitempty"` // Created Time
	UpdateAt *gtime.Time `json:"update_at,omitempty"` // Updated Time
	DeletedAt *gtime.Time `json:"deleted_at,omitempty"` // Deleted Time
}

