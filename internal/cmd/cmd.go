package cmd

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/os/gcmd"
	"proxyServer/api/ws"
	"proxyServer/internal/consts"
	cipc "proxyServer/internal/controller/ipc"

	"proxyServer/internal/controller/hello"

	"proxyServer/internal/controller/user"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			s.SetPort(8000)
			s.Group("/", func(group *ghttp.RouterGroup) {

				hub := ws.NewHub()

				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					hello.New(),
					user.New(),
					cipc.New(hub),
					//ws.New(),
				)

				go hub.Run()
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
