package main

import (
	"proxyServer/internal/cmd"
	_ "proxyServer/internal/logic"
	"proxyServer/internal/packed"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	packed.InitRedis()
	cmd.Main.Run(gctx.GetInitCtx())

}
