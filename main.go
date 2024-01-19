package main

import (
	_ "github.com/BioforestChain/dweb-browser-proxy/app/offline_storage/logic"
	_ "github.com/BioforestChain/dweb-browser-proxy/app/pubsub/logic"
	"github.com/BioforestChain/dweb-browser-proxy/cmd/server"
	_ "github.com/BioforestChain/dweb-browser-proxy/internal/logic"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/mongodb"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/redis"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	redis.InitRedis()
	mongodb.Init()
	server.Main.Run(gctx.GetInitCtx())
}
