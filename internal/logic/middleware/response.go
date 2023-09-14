// 统一响应处理
package middleware

import (
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"net/http"
	"proxyServer/internal/packed"
)

type Response struct {
	Code    int         `json:"code"    dc:"业务码"`
	Message string      `json:"message" dc:"业务码说明"`
	Data    interface{} `json:"data"    dc:"返回的数据"`
}

func ResponseHandler(r *ghttp.Request) {
	r.Middleware.Next()
	if r.Response.BufferLength() > 0 {
		return
	}

	var (
		res  = r.GetHandlerResponse()
		err  = r.GetError()
		code = gerror.Code(err)
		msg  string
	)
	if err != nil {
		if code == gcode.CodeNil {
			code = gcode.CodeInternalError
		}
		msg = err.Error()
	} else {
		if r.Response.Status > 0 && r.Response.Status != http.StatusOK {
			msg = http.StatusText(r.Response.Status)
			switch r.Response.Status {
			case http.StatusNotFound:
				code = gcode.CodeNotFound
			case http.StatusForbidden:
				code = gcode.CodeNotAuthorized
			default:
				code = gcode.CodeUnknown
			}
			// It creates error as it can be retrieved by other middlewares.
			err = gerror.NewCode(code, msg)
			r.SetError(err)
		} else {
			code = gcode.CodeOK
			msg = packed.Err.GetErrorMessage(code.Code())
		}
	}
	r.Response.WriteJson(Response{
		Code:    code.Code(),
		Message: msg,
		Data:    res,
	})
}
