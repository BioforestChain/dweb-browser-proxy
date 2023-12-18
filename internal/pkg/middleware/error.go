package middleware

import (
	error2 "github.com/BioforestChain/dweb-browser-proxy/internal/pkg/error"
	"github.com/gogf/gf/v2/net/ghttp"
	"net/http"
)

func ErrorHandler(r *ghttp.Request) {
	r.Middleware.Next()
	if r.Response.Status >= http.StatusInternalServerError && r.Response.Status < error2.InitRedisErr {
		//if r.Response.Status >= http.StatusInternalServerError {
		r.Response.WriteStatus(http.StatusInternalServerError)
		r.Response.ClearBuffer()
		r.Response.Write(Response{http.StatusInternalServerError, "The server is busy, please try again later!", nil})
	}
}
