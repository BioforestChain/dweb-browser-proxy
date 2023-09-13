package cmd

import (
	"context"
	"errors"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/os/gcmd"
	"golang.org/x/time/rate"
	"io"
	"log"
	"net/http"
	v1 "proxyServer/api/client/v1"
	"proxyServer/api/ws"
	"proxyServer/internal/consts"
	"proxyServer/internal/controller/auth"
	"proxyServer/internal/controller/ping"
	"proxyServer/internal/controller/pre_user"
	"proxyServer/internal/controller/user"
	"proxyServer/internal/logic/middleware"
	"proxyServer/ipc"
	"strings"
	"time"
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

func MiddlewareLimitHandler(r *ghttp.Request, limit *rate.Limiter, clientId string) bool {
	r.Middleware.Next()
	// 请求限制器,如果限制成功则处理请求
	if clientId == "8cb46dde8d8edb41994e0b88f87a31dc" && !limit.Allow() {
		r.Response.WriteStatus(http.StatusTooManyRequests)
		r.Response.ClearBuffer()
		r.Response.Writeln("哎哟请求过快，服务器居然需要休息下，请稍后再试吧！")
		return false
	}
	return true
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

			s.Group("/", func(group *ghttp.RouterGroup) {

				// 创建一个每 200 毫秒限1个请求的限制器
				limitNum, _ := g.Cfg().Get(context.Background(), "rate_limiter.limit")
				limitNumDur := limitNum.Duration() * time.Millisecond
				burst, _ := g.Cfg().Get(context.Background(), "rate_limiter.burst")
				limit := rate.NewLimiter(rate.Every(limitNumDur), burst.Int())

				group.Middleware(
					//MiddlewareLimitHandler(limit),
					MiddlewareErrorHandler,
				)
				group.ALL("/*any", func(r *ghttp.Request) {
					req := &v1.IpcReq{}
					req.Header = strings.Join(r.Header["Content-Type"], "")
					req.Method = r.Method
					req.URL = r.GetUrl()
					req.Host = r.GetHost()
					//TODO 暂定用 query 参数传递
					req.ClientID = r.Get("clientID").String()
					rateRes := MiddlewareLimitHandler(r, limit, req.ClientID)
					if rateRes {
						resIpc, err := Proxy2Ipc(ctx, hub, req)
						if err != nil {
							g.Log().Warning(ctx, "Proxy2Ipc err :", err)
						}

						for k, v := range resIpc.Header {
							r.Response.Header().Set(k, v)
						}

						if _, err = io.Copy(r.Response.Writer, resIpc.Body); err != nil {
							// TODO
							r.Response.WriteStatus(400, "请求出错")
						}
					}
				})
			})

			s.Group("/proxy", func(group *ghttp.RouterGroup) {

				group.Middleware(
					ghttp.MiddlewareHandlerResponse,
					ghttp.MiddlewareCORS,
					MiddlewareErrorHandler,
				)
				group.Group("/", func(group *ghttp.RouterGroup) {
					group.Bind(
						//排除不受JWT认证的路由
						ping.New(),
						auth.New(),
						pre_user.New(),
					)
					group.Middleware(middleware.JWTAuth)
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

func Proxy2Ipc(ctx context.Context, hub *ws.Hub, req *v1.IpcReq) (res *ipc.Response, err error) {
	// 验证 req.Host 是否存于数据库中
	//valCheckUrl := service.User().IsDomainExist(ctx, model.CheckUrlInput{Host: req.Host})
	//if !valCheckUrl {
	//	//抱歉，您的域名尚未注册
	//	res.Ipc = fmt.Sprintf(`{"msg": "%s"}`, gerror.Newf(`Sorry, your domain name "%s" is not registered yet`, req.Host))
	//	return res, nil
	//}
	client := hub.GetClient(req.ClientID)
	if client == nil {
		return nil, errors.New("the service is unavailable")
	}
	clientIpc := client.GetIpc()

	reqIpc := clientIpc.Request(req.URL, ipc.RequestArgs{
		Method: req.Method,
		Header: map[string]string{"Content-Type": req.Header},
	})
	resIpc, err := clientIpc.Send(ctx, reqIpc)
	if err != nil {
		log.Println("ipc response err: ", err)
		return nil, err
	}

	return resIpc, nil
}
