package cmd

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/os/gcmd"
	"golang.org/x/time/rate"
	"io"
	"net/http"
	v1 "proxyServer/api/client/v1"
	"proxyServer/internal/consts"
	"proxyServer/internal/controller/app"
	"proxyServer/internal/controller/auth"
	"proxyServer/internal/controller/chat"
	"proxyServer/internal/controller/ping"
	"proxyServer/internal/controller/pre_user"
	//"proxyServer/internal/logic/net"
	"proxyServer/internal/controller/net"
	helperIPC "proxyServer/internal/helper/ipc"
	"proxyServer/internal/logic/middleware"
	"proxyServer/internal/packed"
	ws "proxyServer/internal/service/ws"
	"sync"
	"time"
)

func MiddlewareLimitHandler() func(r *ghttp.Request) {
	s := sync.Map{}
	return func(r *ghttp.Request) {
		r.Middleware.Next()
		clientID := r.Get("client_id").String()
		v, ok := s.Load(clientID)
		if !ok {
			var limit *rate.Limiter
			// set limit burst
			// default: 100ms 1 burst ; manifest/config/config.yaml
			limitNum, _ := g.Cfg().Get(context.Background(), "rateLimiter.limit")
			limitNumDur := limitNum.Duration() * time.Millisecond
			burst, _ := g.Cfg().Get(context.Background(), "rateLimiter.burst")
			limit = rate.NewLimiter(rate.Every(limitNumDur), burst.Int())
			s.Store(clientID, limit)
			v = limit
		}
		tmp := v.(*rate.Limiter)
		// Request a limiter, which processes the request if throttling succeeds
		if !tmp.Allow() {
			r.Response.WriteStatus(http.StatusTooManyRequests)
			r.Response.ClearBuffer()
			r.Response.Write(middleware.Response{http.StatusTooManyRequests, "The request is too fast, please try again later!", nil})
		}
	}
}
func MiddlewareErrorHandler(r *ghttp.Request) {
	r.Middleware.Next()
	if r.Response.Status >= http.StatusInternalServerError && r.Response.Status < consts.InitRedisErr {
		r.Response.WriteStatus(http.StatusInternalServerError)
		r.Response.ClearBuffer()
		r.Response.Write(middleware.Response{http.StatusInternalServerError, "The server is busy, please try again later!", nil})
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
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(
					MiddlewareLimitHandler(),
				)
				group.ALL("/*any", func(r *ghttp.Request) {

					req := &v1.IpcReq{}
					//req.Header = strings.Join(r.Header["Content-Type"], "")
					req.Header = r.Header
					req.Method = r.Method
					req.URL = r.GetUrl()
					req.Host = r.GetHost()
					// TODO 需要优化body https://medium.com/@owlwalks/sending-big-file-with-minimal-memory-in-golang-8f3fc280d2c
					req.Body = r.GetBody()
					//TODO 暂定用 query 参数传递
					req.ClientID = req.Host
					resIpc, err := packed.Proxy2Ipc(ctx, hub, req)
					if err != nil {
						resIpc = packed.IpcErrResponse(consts.ServiceIsUnavailable, err.Error())
					}
					for k, v := range resIpc.Header {
						r.Response.Header().Set(k, v)
					}
					bodyStream := resIpc.Body.Stream()
					if bodyStream == nil {
						if _, err = io.Copy(r.Response.Writer, resIpc.Body); err != nil {
							r.Response.WriteStatus(400, "请求出错")
						}
					} else {
						data, err := helperIPC.ReadStreamWithTimeout(bodyStream, 10*time.Second)
						if err != nil {
							r.Response.WriteStatus(400, err)
						} else {
							r.Response.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
							_, _ = r.Response.Writer.Write(data)
						}
					}
				})
			})

			s.Group("/proxy", func(group *ghttp.RouterGroup) {
				group.Middleware(
					middleware.ResponseHandler,
					ghttp.MiddlewareCORS,
					MiddlewareErrorHandler,
				)
				group.Group("/", func(group *ghttp.RouterGroup) {
					group.Bind(
						//Exclude routes that are not JWT certified
						ping.New(),
						auth.New(),
						pre_user.New(),
					)
					//group.Middleware(middleware.JWTAuth)
					group.Bind(
						net.New(),
						app.New(),
						chat.New(hub),
					)
				})
				group.GET("/ws", func(r *ghttp.Request) {
					//
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
