package v1

import "github.com/gogf/gf/v2/frame/g"

type WsReq struct {
	g.Meta `path:"/ws/test" tags:"WsService" method:"get" summary:"ws"`
	//Name                 string `v:"required"` //用户名，昵称？
	//PublicKey            string `v:"required"`
	//DeviceIdentification string `v:"required"` // imei码，身份标识
	//Remark               string
}
type WsRes struct {
	Msg string
	//Name                 string `v:"required"` //用户名，昵称？
	//PublicKey            string `v:"required"`
	//DeviceIdentification string `v:"required"` // imei码，身份标识
	//Remark               string
}
