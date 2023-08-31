package cmd

import (
	"context"
	"encoding/json"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/os/gcmd"
	"log"
	"net/http"
	v1 "proxyServer/api/client/v1"
	"proxyServer/api/ws"
	"proxyServer/internal/consts"
	"proxyServer/internal/controller/auth"
	"proxyServer/internal/controller/ping"
	"proxyServer/internal/controller/user"
	"proxyServer/internal/logic/middleware"
	"proxyServer/ipc"
	"strings"
)

func MiddlewareAuth(r *ghttp.Request) {
	token := r.Get("token")

	if token.String() == "123456" {
		//r.Response.Writeln("auth")

		r.Middleware.Next()
	} else {
		r.Response.WriteStatus(http.StatusForbidden)

	}
}

func MiddlewareErrorHandler(r *ghttp.Request) {
	r.Middleware.Next()
	if r.Response.Status >= http.StatusInternalServerError {
		r.Response.ClearBuffer()
		r.Response.Writeln("哎哟我去，服务器居然开小差了，请稍后再试吧！")
	}
}

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()

			//s.SetRouteOverWrite(true)
			hub := ws.NewHub()
			go hub.Run()
			s.BindHandler("/*any", func(r *ghttp.Request) {
				var (
					req *v1.IpcReq
					res *v1.IpcRes
					err error
				)
				req = &v1.IpcReq{}
				req.Header = strings.Join(r.Header["Content-Type"], "")
				req.Method = r.Method
				req.URL = r.GetUrl()
				req.Host = r.GetHost()
				//TODO 暂定用 query 参数传递
				req.ClientID = r.Get("clientID").String()
				res, err = Proxy2Ipc(ctx, hub, req)
				if err != nil {
					g.Log().Warning(ctx, "Proxy2Ipc err :", err)
				}
				r.Response.Write(res)
			})

			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(
					ghttp.MiddlewareHandlerResponse,
					ghttp.MiddlewareCORS,
					MiddlewareErrorHandler,
				)
				group.Group("/", func(group *ghttp.RouterGroup) {
					group.Bind(
						ping.New(),
						auth.New(),
					)
					group.Middleware(middleware.JWTAuthMiddleware)
					//group.Middleware(service.Middleware().Auth)
					group.Bind(
						user.New(),
					)
				})
				s.BindHandler("/ws", func(r *ghttp.Request) {
					ws.ServeWs(hub, r.Response.Writer, r.Request)
				})

				// Special handler that needs authentication.
				//group.Group("/", func(group *ghttp.RouterGroup) {
				//	group.Middleware(service.Middleware().Auth)
				//	group.ALLMap(g.Map{
				//		"/user/profile": user.New().Profile,
				//	})
				//})

				//group.Group("/user", func(group *ghttp.RouterGroup) {
				//	group.GET("/client-list", func(r *ghttp.Request) {
				//
				//		r.Response.Write("info")
				//
				//	})
				//	group.POST("/edit", func(r *ghttp.Request) {
				//		r.Response.Write("edit")
				//	})
				//	group.DELETE("/drop", func(r *ghttp.Request) {
				//		r.Response.Write("drop")
				//	})
				//})

			})

			enhanceOpenAPIDoc(s)
			s.Run()
			return nil
		},
	}
)

func enhanceOpenAPIDoc(s *ghttp.Server) {
	openapi := s.GetOpenApi()
	openapi.Config.CommonResponse = ghttp.DefaultHandlerResponse{}
	openapi.Config.CommonResponseDataField = `Data`

	// API description.
	openapi.Info = goai.Info{
		Title:       consts.OpenAPITitle,
		Description: consts.OpenAPIDescription,
		Contact: &goai.Contact{
			Name: "GoFrame",
			URL:  "https://goframe.org",
		},
	}
}

// Proxy2Ipc
//
//	@Description: 转到Ipc
//	@param ctx
//	@param hub
//	@param req
//	@return res
//	@return err

func Proxy2Ipc(ctx context.Context, hub *ws.Hub, req *v1.IpcReq) (res *v1.IpcRes, err error) {
	req = &v1.IpcReq{
		Host:     req.Host,
		Method:   req.Method,
		URL:      req.URL,
		Header:   req.Header,
		ClientID: req.ClientID,
	}
	res = &v1.IpcRes{}
	// 验证 req.Host 是否存于数据库中
	//valCheckUrl := service.User().IsDomainExist(ctx, model.CheckUrlInput{Host: req.Host})
	//if !valCheckUrl {
	//	//抱歉，您的域名尚未注册
	//	res.Ipc = fmt.Sprintf(`{"msg": "%s"}`, gerror.Newf(`Sorry, your domain name "%s" is not registered yet`, req.Host))
	//	return res, nil
	//}
	client := hub.GetClient(req.ClientID)
	if client == nil {
		res.Ipc = "The service is unavailable"
		return res, nil
	}
	clientIpc := client.GetIpc()

	reqIpc := clientIpc.Request(req.URL, ipc.RequestArgs{
		Method: req.Method,
		Header: map[string]string{"Content-Type": req.Header},
	})
	resIpc, err := clientIpc.Send(ctx, reqIpc)
	//fmt.Printf("------------resIpc:%#v\n", resIpc)
	if err != nil {
		log.Println("ipc response err: ", err)
		//res.Ipc = fmt.Sprintf(`{"msg": "%s"}`, err.Error())
		res.Ipc = err.Error()
		return res, err
	}
	//todo
	//for k, v := range resIpc.Header {
	//	fmt.Printf("-----------k:%#v\n", k)
	//	fmt.Printf("-----------v:%#v\n", v)
	//	//w.Header().Set(k, v)
	//}

	resStr, err := json.Marshal(resIpc)
	if err != nil {
		//res.Ipc = fmt.Sprintf(`{"msg": "%s"}`, err.Error())
		res.Ipc = err.Error()
		return res, err
	}
	res.Ipc = string(resStr)
	return res, err
}
