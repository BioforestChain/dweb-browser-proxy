package v1

// IpcReq
// @Description:
type IpcReq struct {
	//g.Meta `path:"/ipc/test" tags:"IpcTestService" method:"get" summary:"test a new ipc"`
	Header   string `v:"required"` //
	Method   string `v:"required"` //
	URL      string `v:"required"` //
	Host     string `v:"required"` //
	ClientID string `v:"required"` // 用户id
}

//type IpcRes struct {
//
//	//Ipc string `json:"ipc"`
//}
