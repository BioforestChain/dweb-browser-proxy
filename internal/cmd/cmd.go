package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/gogf/gf/v2/errors/gerror"
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
	"proxyServer/internal/model"
	"proxyServer/internal/service"
	"proxyServer/ipc"
	"strings"
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
			limitNum, _ := g.Cfg().Get(context.Background(), "rate_limiter.limit")
			limitNumDur := limitNum.Duration() * time.Millisecond
			burst, _ := g.Cfg().Get(context.Background(), "rate_limiter.burst")
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
					req.Header = strings.Join(r.Header["Content-Type"], "")
					req.Method = r.Method
					req.URL = r.GetUrl()
					req.Host = r.GetHost()
					//TODO 暂定用 query 参数传递
					req.ClientID = r.Get("client_id").String()
					resIpc, err := Proxy2Ipc(ctx, hub, req)
					if err != nil {
						resIpc = ipcErrResponse(consts.ServiceIsUnavailable, err.Error())
					}
					for k, v := range resIpc.Header {
						r.Response.Header().Set(k, v)
					}
					if _, err = io.Copy(r.Response.Writer, resIpc.Body); err != nil {
						r.Response.WriteStatus(400, "请求出错")
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
					group.Middleware(middleware.JWTAuth)
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
//	@Description: The request goes to the IPC object for processing
//	@param ctx
//	@param hub
//	@param req
//	@return res
//	@return err
func Proxy2Ipc(ctx context.Context, hub *ws.Hub, req *v1.IpcReq) (res *ipc.Response, err error) {
	client := hub.GetClient(req.ClientID)
	if client == nil {
		return nil, errors.New("the service is unavailable")
	}
	// Verify req.Host exists in the database
	valCheckUrl := service.User().IsDomainExist(ctx, model.CheckUrlInput{Host: req.Host})
	if !valCheckUrl {
		log.Println(gerror.Newf(`Sorry, your domain name "%s" is not registered yet`, req.Host))
		return nil, gerror.Newf(`Sorry, your domain name "%s" is not registered yet`, req.Host)
	}
	// Verify req.ClientID exists in the database
	valCheckUser := service.User().IsUserExist(ctx, model.CheckUserInput{UserIdentification: req.ClientID})
	if !valCheckUser {
		log.Println(gerror.Newf(`Sorry, your user "%s" is not registered yet`, req.ClientID))
		return nil, gerror.Newf(`Sorry, your user "%s" is not registered yet`, req.ClientID)
	}
	clientIpc := client.GetIpc()
	reqIpc := clientIpc.Request(req.URL, ipc.RequestArgs{
		Method: req.Method,
		Header: map[string]string{"Content-Type": req.Header},
	})
	resIpc, err := clientIpc.Send(ctx, reqIpc)
	if err != nil {
		return nil, err
	}
	return resIpc, nil
}
func ipcErrResponse(code int, msg string) *ipc.Response {
	body := fmt.Sprintf(`{"code": %d, "message": %s, "data": nil}`, code, msg)
	res := ipc.NewResponse(
		1,
		400,
		ipc.NewHeaderWithExtra(map[string]string{
			"Content-Type": "application/json",
		}),
		ipc.NewBodySender([]byte(body), nil),
		nil,
	)
	return res
}
