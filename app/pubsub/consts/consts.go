package consts

const RedisPrefix = "proxy_net:"
const FormatKey = "%s%s"

const (
	XDwebHostMMID   = "X-Dweb-Host"
	XDwebHostDomain = "X-Dweb-Host-Domain"

	PubsubAppMMID = "X-Dweb-Pubsub-App"
	PubsubMMID    = "X-Dweb-Pubsub"
	// 权限类型: 0:无认证，1:acl，2:基于密码，3:基于角色，4:etc
	PubsubPermissionTypeAcl = 1 //
)
