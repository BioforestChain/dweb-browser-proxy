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

var mySecret = []byte(consts.JwtSecret)

func keyFunc(_ *jwt.Token) (i interface{}, err error) {
	return mySecret, nil
}

// JWTAuthMiddleware 基于JWT的认证中间件
func JWTAuth(r *ghttp.Request) {
	s := new(sMiddleware)
	//排除不受JWT认证的路由
	if r.URL.Path == consts.RouteProxyUserClientReg {
		r.Middleware.Next()
	} else {
		// 客户端携带Token 放在请求头
		token := r.Header.Get("Authorization")
		if len(token) == 0 {
			r.Response.WriteJson(Response{
				Code:    consts.TokenCannotBeNull,
				Message: packed.Err.GetErrorMessage(consts.TokenCannotBeNull),
				Data:    nil,
			})
			r.Exit()
			return
		}
		mc, err := s.ParseToken(token)
		log.Println("jwtAccessToken err is ", err)
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
}

func (s *sMiddleware) GenToken(userID uint32) (string, string, int64, error) {
	jwtExpire, _ := g.Cfg().Get(s.Ctx, "auth.jwt_token_expire")
	JwtRefreshTokenExpire, _ := g.Cfg().Get(s.Ctx, "auth.jwt_refresh_token_expire")
	jwtExpireInt64 := jwtExpire.Int64()
	JwtRefreshTokenExpireInt64 := JwtRefreshTokenExpire.Int64()

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
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(mySecret)

	log.Println("AccessToken ExpiresAt:", jwt.NewNumericDate(time.Now().Add(time.Hour*time.Duration(jwtExpireInt64))))
	// 使用指定的secret签名并获得完整的编码后的字符串token
	rc := entity.MyClaims{
		userID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(JwtRefreshTokenExpireInt64))), // 过期时间
			NotBefore: jwt.NewNumericDate(time.Now()),                                                            // 签发时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                                                            // 生效时间
			Issuer:    "jwt",                                                                                     // 签发人
		},
	}
	refreshToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, rc).SignedString(mySecret)

	return token, refreshToken, c.ExpiresAt.Unix(), nil
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
// 第一步 : 判断 refreshToken 格式对的，没有过期的
// 第二步 : 判断 accessToken 格式对的，但是是过期的
// 第三步 : 生成双 token
func (s *sMiddleware) RefreshToken(accessToken, refreshToken string) (string, string, int64, uint32, error) {
	// refresh token无效直接返回
	if _, err := jwt.Parse(refreshToken, keyFunc); err != nil {
		return "", "", 0, 0, err
	}

	// 从旧access token中解析出claims数据
	var claims entity.MyClaims
	_, err := jwt.ParseWithClaims(accessToken, &claims, keyFunc)
	fmt.Printf("claims %#v\n: ", claims)
	v, _ := err.(*jwt.ValidationError)
	// 当access token是过期错误 并且 refresh token没有过期时就创建一个新的access token
	if v.Errors == jwt.ValidationErrorExpired {
		token, refreshToken, expiresAt, _ := s.GenToken(claims.UserID)
		return token, refreshToken, expiresAt, claims.UserID, nil
	}
	return "", "", 0, 0, err
}
