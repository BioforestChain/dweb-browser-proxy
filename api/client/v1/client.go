package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type ClientRegReq struct {
	g.Meta         `path:"/user/client-reg" tags:"ClientRegService" method:"post" summary:"Sign up a new client"`
	Name           string
	Identification string
	Remark         string
}
type ClientRegRes struct {
}
