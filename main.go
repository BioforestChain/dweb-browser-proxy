package main

import (
	"github.com/BioforestChain/dweb-browser-proxy/internal/cmd/server"
	_ "github.com/BioforestChain/dweb-browser-proxy/internal/logic"
	"github.com/BioforestChain/dweb-browser-proxy/internal/pkg/redis"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	redis.InitRedis()
	server.Main.Run(gctx.GetInitCtx())
}
