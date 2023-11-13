package main

import (
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/os/gctx"
	"proxyServer/internal/cmd"
	_ "proxyServer/internal/logic"
	"proxyServer/internal/packed"
)

func main() {
	packed.InitRedis()
	cmd.Main.Run(gctx.GetInitCtx())
}
