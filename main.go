package main

import (
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/os/gctx"
	_ "proxyServer/internal/packed"

	"proxyServer/internal/cmd"
	_ "proxyServer/internal/logic"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
