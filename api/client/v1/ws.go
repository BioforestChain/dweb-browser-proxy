package v1

import "github.com/gogf/gf/v2/frame/g"

type WsReq struct {
	g.Meta `path:"/ws/test" tags:"WsService" method:"get" summary:"ws"`
}
type WsRes struct {
	Msg string
}
