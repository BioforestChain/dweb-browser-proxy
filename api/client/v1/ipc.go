package v1

type IpcTestReq struct {
	//g.Meta `path:"/ipc/test" tags:"IpcTestService" method:"get" summary:"test a new ipc"`
	Header string `v:"required"` //
	Method string `v:"required"`
	URL    string `v:"required"` //
	Host   string //
	//Remark               string
}
type IpcTestRes struct {
	Ipc string `json:"ipc"`
}
