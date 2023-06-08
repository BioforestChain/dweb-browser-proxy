package main

import (
	_ "frpConfManagement/internal/packed"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/os/gctx"

	"frpConfManagement/internal/cmd"
	_ "frpConfManagement/internal/logic"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
