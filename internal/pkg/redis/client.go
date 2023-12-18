package redis

import (
	"context"
	"encoding/json"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"strconv"
)

type RedisClient struct {
	Client *redis.Client
}

var Ctx context.Context

func InitRedis() {
	instances, err := InitRedisPool("default")
	if err == nil {
		SetRedisInstance("default", instances)
	}
	//go closePool()
}

func InitRedisPool(dbName string) ([]RedisInstance, error) {
	maxActive, _ := g.Cfg().Get(Ctx, "redis."+dbName+".maxActive")
	getMaxActive, _ := strconv.Atoi(maxActive.Val().(string))
	servers, err := g.Cfg().Get(Ctx, "redis."+dbName+".servers")
	serversArr := servers.Array()
	if err != nil {
		return nil, err
	}
	redisPools := make([]RedisInstance, len(serversArr))
	for k, v := range serversArr {
		server, ok := v.(map[string]interface{})
		if !ok {
			log.Printf("parse server config failed, index=%d", k)
			continue
		}
		address := server["address"].(string)
		dbNum := server["db"].(json.Number)
		dbNumStr := string(dbNum)
		db, _ := strconv.Atoi(dbNumStr)
		passwd := server["pass"].(string)
		if len(address) == 0 {
			os.Exit(1)
		}
		// 建立连接
		rds := &RedisClient{}
		// 使用默认的 context
		// 使用 redis 库里的 NewClient 初始化连接
		rds.Client = redis.NewClient(&redis.Options{
			Addr:     address,
			Password: passwd,
			DB:       db,
			PoolSize: getMaxActive,
			//IdleTimeout: time.Duration(idleTimeout) * time.Second,
		})
		redisPools[k] = RedisInstance{
			Client: rds.Client,
		}
	}
	return redisPools, nil
}
