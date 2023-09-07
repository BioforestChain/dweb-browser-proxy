package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type Req struct {
	g.Meta `path:"/ping" tags:"Ping" method:"get" summary:"You first ping api"`
}
type Res struct {
	g.Meta `mime:"text/html" example:"string"`
}
