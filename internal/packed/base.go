package packed

import (
	"context"
	"github.com/gogf/gf/v2/os/gctx"
)

// 资源释放
var CtxSrcRelease, CancelSrcRelease = context.WithCancel(gctx.New())
