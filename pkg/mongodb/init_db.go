package mongodb

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Database struct {
	Mongo *mongo.Client
}

var DB *Database

// Init
//
//	@Description: 初始化
func Init() {
	DB = &Database{
		Mongo: SetConnect(),
	}
}

// SetConnect 设置客户端连接配置
//
//	@Description: uri := "mongodb://localhost:27017"
//
// @return *mongo.Client
func SetConnect() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	uriInit, _ := g.Cfg().Get(ctx, "database.mongodb.link")
	uri := uriInit.String()
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetMaxPoolSize(20)) // 连接池
	if err != nil {
		fmt.Println(err)
	}
	return client
}
