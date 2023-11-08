// ==========================================================================
// Code generated by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// NetDao is the data access object for table net.
type NetDao struct {
	table   string     // table is the underlying table name of the DAO.
	group   string     // group is the database configuration group name of current DAO.
	columns NetColumns // columns contains all the column names of Table for convenient usage.
}

// NetColumns defines and stores column names for table net.
type NetColumns struct {
	Id        string //
	NetId     string // 网络模块id
	Domain    string // 域名
	Remark    string // 备注信息
	Timestamp string // 时间戳
	Port      string // 端口
	IsOnline  string // 上线：1上线，0下线
	CreatedAt string // Created Time
	UpdateAt  string // Updated Time
	DeletedAt string // Deleted Time
}

// netColumns holds the columns for table net.
var netColumns = NetColumns{
	Id:        "id",
	NetId:     "net_id",
	Domain:    "domain",
	Remark:    "remark",
	Timestamp: "timestamp",
	Port:      "port",
	IsOnline:  "is_online",
	CreatedAt: "created_at",
	UpdateAt:  "update_at",
	DeletedAt: "deleted_at",
}

// NewNetDao creates and returns a new DAO object for table data access.
func NewNetDao() *NetDao {
	return &NetDao{
		group:   "default",
		table:   "net",
		columns: netColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *NetDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *NetDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *NetDao) Columns() NetColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *NetDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *NetDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *NetDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}