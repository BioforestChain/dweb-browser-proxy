package server

import (
	"context"
	"github.com/BioforestChain/dweb-browser-proxy/app/offline_storage/controller/offline_msg"
	pubsubCtrl "github.com/BioforestChain/dweb-browser-proxy/app/pubsub/controller"
	"github.com/BioforestChain/dweb-browser-proxy/app/pubsub/controller/pubsub_permission"
	"github.com/BioforestChain/dweb-browser-proxy/internal/consts"
	"github.com/BioforestChain/dweb-browser-proxy/internal/controller"
	"github.com/BioforestChain/dweb-browser-proxy/internal/controller/app"
	"github.com/BioforestChain/dweb-browser-proxy/internal/controller/net"
	"github.com/BioforestChain/dweb-browser-proxy/internal/controller/ping"
	"github.com/BioforestChain/dweb-browser-proxy/internal/controller/pre_user"
	middleware2 "github.com/BioforestChain/dweb-browser-proxy/pkg/middleware"
	ws2 "github.com/BioforestChain/dweb-browser-proxy/pkg/ws"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/os/gcmd"
)

var Main = gcmd.Command{
	Name:  "main",
	Usage: "main",
	Brief: "start http server",
	Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
		s := g.Server()
		//s.SetRouteOverWrite(true)
		hub := ws2.NewHub()
		go hub.Run()

		s.Group("/", func(group *ghttp.RouterGroup) {
			group.Middleware(middleware2.LimitHandler())
			group.ALL("/*any", func(r *ghttp.Request) {
				//TODO offline msg by post base on https
				controller.Proxy.Forward(ctx, r, hub)
			})
		})

		s.Group("/proxy", func(group *ghttp.RouterGroup) {
			group.Middleware(
				middleware2.ResponseHandler,
				ghttp.MiddlewareCORS,
				middleware2.ErrorHandler,
			)
			group.Group("/", func(group *ghttp.RouterGroup) {
				group.Bind(
					ping.New(),
					//auth.New(),
					pre_user.New(),
					net.New(),
					app.New(),

					pubsub_permission.New(),
					pubsubCtrl.NewChat(hub),
					offline_msg.New(),
				)
			})

			group.GET("/ws", func(r *ghttp.Request) {
				controller.WsIns.Connect(hub, r.Response.Writer, r.Request)
			})

			group.POST("/pubsub/test/pub", func(r *ghttp.Request) {
				pubsubCtrl.PubSub.Pub(ctx, hub, r.Response.Writer, r)
			})

			group.POST("/pubsub/test/sub", func(r *ghttp.Request) {
				pubsubCtrl.PubSub.Sub(ctx, hub, r.Response.Writer, r)
			})
		})
		enhanceOpenAPIDoc(s)
		s.Run()
		return nil
	},
}

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