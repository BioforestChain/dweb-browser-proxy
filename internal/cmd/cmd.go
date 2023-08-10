package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/os/gcmd"
	"log"
	"net/http"
	v1 "proxyServer/api/client/v1"
	"proxyServer/api/ws"
	"proxyServer/internal/consts"
	"proxyServer/internal/controller/hello"
	"proxyServer/internal/controller/user"
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

func MiddlewareCORS(r *ghttp.Request) {
	//r.Response.Writeln("cors")
	r.Response.CORSDefault()
	r.Middleware.Next()
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
			var req v1.IpcTestReq
			s.BindHandler("/*any", func(r *ghttp.Request) {

				var (
					res *v1.IpcTestRes
					err error
				)
				req.Header = strings.Join(r.Header["Content-Type"], "")
				req.Method = r.Method
				req.URL = GetURL(r)
				res, err = Proxy2Ipc(ctx, hub, req)
				if err != nil {
					log.Fatalln("Proxy2Ipc err is ", err)
				}
				r.Response.WriteJson(res)
			})

			s.Group("/", func(group *ghttp.RouterGroup) {

				group.Middleware(ghttp.MiddlewareHandlerResponse, MiddlewareCORS)
				group.Group("/", func(group *ghttp.RouterGroup) {
					group.Bind(
						user.New(),
						hello.New(),
					)
				})
				s.BindHandler("/ws", func(r *ghttp.Request) {
					ws.ServeWs(hub, r.Response.Writer, r.Request)
				})

				group.Middleware(MiddlewareAuth, MiddlewareErrorHandler)

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

func GetURL(r *ghttp.Request) (Url string) {
	scheme := "http://"
	if r.TLS != nil {
		scheme = "https://"
	}
	return strings.Join([]string{scheme, r.Host, r.RequestURI}, "")
}

func Proxy2Ipc(ctx context.Context, hub *ws.Hub, req v1.IpcTestReq) (res *v1.IpcTestRes, err error) {
	res = &v1.IpcTestRes{}
	client := hub.GetClient("test")
	//fmt.Printf("client: %#v\n", client)
	if client == nil {
		res.Ipc = "The service is unavailable"
		return res, nil
	}
	clientIpc := client.GetIpc()

	reqIpc := clientIpc.Request(req.URL, ipc.RequestArgs{
		Method: req.Method,
		Header: map[string]string{"Content-Type": req.Header},
	})
	resIpc, err := clientIpc.Send(reqIpc)
	fmt.Printf("------------resIpc", resIpc)
	if err != nil {
		log.Println("ipc response err: ", err)
		//res.Ipc = fmt.Sprintf(`{"msg": "%s"}`, err.Error())
		res.Ipc = err.Error()
		return res, err
	}
	//todo

	//for k, v := range resIpc.Header {
	//	fmt.Printf("------------k", k)
	//	fmt.Printf("-----------v", v)
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
