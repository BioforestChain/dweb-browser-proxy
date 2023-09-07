package middleware

import (
	"errors"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/golang-jwt/jwt/v4"
	"log"
	"proxyServer/internal/consts"
	"proxyServer/internal/model/entity"
	"proxyServer/internal/packed"
	"time"
)

var mySecret = []byte("jwt")

func keyFunc(_ *jwt.Token) (i interface{}, err error) {
	return mySecret, nil
}

// JWTAuthMiddleware 基于JWT的认证中间件
func JWTAuth(r *ghttp.Request) {
	s := new(sMiddleware)
	// 客户端携带Token有三种方式 1.放在请求头 2.放在请求体 3.放在URI
	token := r.Header.Get("Authorization") // 可以获取路径中的 name 参数
	fmt.Println("token", token)
	if len(token) == 0 {
		r.Response.WriteJson(Response{
			Code:    consts.TokenCannotBeNull,
			Message: packed.Err.GetErrorMessage(consts.TokenCannotBeNull),
			Data:    nil,
		})
		r.Exit() // 中止
		return
	}
	mc, err := s.ParseToken(token)
	log.Println("jwtService err is ", err)
	if err != nil {
		r.Response.WriteJson(Response{
			Code:    consts.TokenIsInvalid,
			Message: packed.Err.GetErrorMessage(consts.TokenIsInvalid),
			Data:    nil,
		})
		r.Exit()
		return
	}
	// 将当前请求的userid信息保存到请求的上下文c上
	r.Header.Set(consts.CtxUserIDKey, string(mc.UserID))
	r.Middleware.Next() // 后续的处理函数可以用过c.Get("username")来获取当前请求的用户信息
}

func (s *sMiddleware) GenToken(userID uint32) (string, error) {
	jwtExpire, _ := g.Cfg().Get(s.Ctx, "auth.jwt_expire")

	jwtExpireInt64 := jwtExpire.Int64()
	c := entity.MyClaims{
		userID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(jwtExpireInt64))), // 过期时间
			NotBefore: jwt.NewNumericDate(time.Now()),                                                // 签发时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                                                // 生效时间
			Issuer:    "jwt",                                                                         // 签发人
		},
	}
	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	log.Println("过期时间:", jwt.NewNumericDate(time.Now().Add(time.Hour*time.Duration(jwtExpireInt64))))
	// 使用指定的secret签名并获得完整的编码后的字符串token
	return token.SignedString(mySecret)
}

// ParseToken 解析JWT
func (s *sMiddleware) ParseToken(tokenString string) (*entity.MyClaims, error) {
	// 解析token
	var mc = new(entity.MyClaims)
	token, err := jwt.ParseWithClaims(tokenString, mc, keyFunc)
	if err != nil {
		return nil, err
	}
	// 校验token
	if token.Valid {
		return mc, nil
	}
	return nil, errors.New("invalid token")
}

// RefreshToken 刷新AccessToken
func (s *sMiddleware) RefreshToken(aToken, rToken string) (newAToken, newRToken string, err error) {
	// refresh token无效直接返回
	if _, err = jwt.Parse(rToken, keyFunc); err != nil {
		return
	}
	// 从旧access token中解析出claims数据
	var claims entity.MyClaims
	_, err = jwt.ParseWithClaims(aToken, &claims, keyFunc)
	v, _ := err.(*jwt.ValidationError)
	// 当access token是过期错误 并且 refresh token没有过期时就创建一个新的access token
	if v.Errors == jwt.ValidationErrorExpired {
		token, _ := s.GenToken(claims.UserID)
		return token, "", nil
	}
	return
}
