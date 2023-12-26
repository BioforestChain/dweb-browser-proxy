package consts

import "time"

const RedisPrefix = "proxy_net:"
const FormatKey = "%s%s"

// paging
const InitLimit = 10
const InitPage = 1

const (
	OpenAPITitle       = `GoFrame Demos`
	OpenAPIDescription = `This is a simple demos HTTP server project that is using GoFrame. Enjoy 💖 `
)

const (
	DefaultDateTime      = "1970-01-01 00:00:00"
	DefaultDateFormatEn  = "02/01/2006 15:04"
	DefaultDateFormatMin = "2006-01-02 15:04"
)

const (
	JwtSecret                = "8ab6c8dee22768da1503351069f032cb" // jwt密匙
	CtxUserIDKey             = "userID"
	CtxUserIdentificationKey = "userIdentification"
	Issuer                   = "jwt"
	JwtTTLTimeUnit           = time.Hour
)

const (
	XDwebHostMMID   = "X-Dweb-Host"
	XDwebHostDomain = "X-Dweb-Host-Domain"

	PubsubAppMMID = "X-Dweb-Pubsub-App"
	PubsubMMID    = "X-Dweb-Pubsub"
	// 权限类型: 0:无认证，1:acl，2:基于密码，3:基于角色，4:etc
	PubsubPermissionTypeAcl = 1 //
)

type User struct {
	UserID uint32 `json:"user_id"`
	Token  string
}
