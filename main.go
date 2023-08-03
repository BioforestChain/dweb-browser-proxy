package main

import (
	_ "proxyServer/internal/packed"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/os/gctx"

	"proxyServer/internal/cmd"
	_ "proxyServer/internal/logic"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
