package v1

import "github.com/gogf/gf/v2/frame/g"

type IpcTestReq struct {
	g.Meta `path:"/ipc/test" tags:"IpcTestService" method:"get" summary:"test a new ipc"`
	//Name                 string `v:"required"` //用户名，昵称？
	//PublicKey            string `v:"required"`
	//DeviceIdentification string `v:"required"` // imei码，身份标识
	//Remark               string
}
type IpcTestRes struct {
	Msg string
	//Name                 string `v:"required"` //用户名，昵称？
	//PublicKey            string `v:"required"`
	//DeviceIdentification string `v:"required"` // imei码，身份标识
	//Remark               string
}
