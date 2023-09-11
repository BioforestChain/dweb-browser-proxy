package consts

import "time"

const (
	JwtSecret      = "8ab6c8dee22768da1503351069f032cb" // jwt密匙
	CtxUserIDKey   = "userID"
	Issuer         = "jwt"
	JwtTTLTimeUnit = time.Hour
)

type User struct {
	UserID uint32 `json:"user_id"`
	Token  string
}
