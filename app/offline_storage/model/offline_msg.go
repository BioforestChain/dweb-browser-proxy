package model

import (
	"github.com/BioforestChain/dweb-browser-proxy/pkg/model"
)

type OfflineMsgDelInput struct {
	Id int64
}

type OfflineMsgListQueryInput struct {
	model.PaginationSearch
	Receiver string `json:"receiver"   in:"query" dc:"接收者"`

	ColName string `json:"colName"   in:"query" dc:"表名"`
	DbName  string `json:"dbName"   in:"query" dc:"数据库名"`
}
