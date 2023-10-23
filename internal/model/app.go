package model

type AppModuleCreateInput struct {
	NetId    string
	AppId    string
	UserName string
	AppName  string
}

type AppModuleInfoCreateInput struct {
	UserName  string
	AppName   string
	AppId     string
	NetId     string
	PublicKey string
	IsInstall uint32
	IsOnline  uint32
	Timestamp string
	Remark    string
}
