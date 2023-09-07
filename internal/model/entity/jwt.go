package entity

import "github.com/golang-jwt/jwt/v4"

// MyClaims 自定义声明结构体并内嵌 jwt.RegisteredClaims
// jwt包自带的jwt.RegisteredClaims 只包含了官方字段，若需要额外记录其他字段，就可以自定义结构体
// 如果想要保存更多信息，都可以添加到这个结构体中
type MyClaims struct {
	UserID uint32 `json:"user_id"`
	jwt.RegisteredClaims
}
