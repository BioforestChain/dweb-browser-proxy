package cmd

import (
	"context"
	"encoding/json"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/os/gcmd"
	"golang.org/x/time/rate"
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
	"proxyServer/internal/packed"
	"proxyServer/ipc"
	"strings"
	"sync"
	"time"
)

func MiddlewareLimitHandler() func(r *ghttp.Request) {
	s := sync.Map{}
	return func(r *ghttp.Request) {
		r.Middleware.Next()
		clientID := r.Get("clientID").String()
		v, ok := s.Load(clientID)
		if !ok {
			var limit *rate.Limiter
			// 创建一个每 200 毫秒限1个请求的限制器
			limitNum, _ := g.Cfg().Get(context.Background(), "rate_limiter.limit")
			limitNumDur := limitNum.Duration() * time.Millisecond
			burst, _ := g.Cfg().Get(context.Background(), "rate_limiter.burst")
			limit = rate.NewLimiter(rate.Every(limitNumDur), burst.Int())
			s.Store(clientID, limit)
			v = limit
		}
		tmp := v.(*rate.Limiter)
		// 请求限制器,如果限制成功则处理请求
		if !tmp.Allow() {
			r.Response.WriteStatus(http.StatusTooManyRequests)
			r.Response.ClearBuffer()
			r.Response.Write(middleware.Response{http.StatusTooManyRequests, "The request is too fast, please try again later!", nil})
		}
	}
}
func MiddlewareErrorHandler(r *ghttp.Request) {
	r.Middleware.Next()
	if r.Response.Status >= http.StatusInternalServerError {
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
					MiddlewareErrorHandler,
				)
				group.ALL("/*any", func(r *ghttp.Request) {
					var (
						req *v1.IpcReq
						res *middleware.Response
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
					r.Response.WriteJsonExit(res)
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
func Proxy2Ipc(ctx context.Context, hub *ws.Hub, req *v1.IpcReq) (res *middleware.Response, err error) {
	req = &v1.IpcReq{
		Host:     req.Host,
		Method:   req.Method,
		URL:      req.URL,
		Header:   req.Header,
		ClientID: req.ClientID,
	}
	res = &middleware.Response{}
	res.Code = consts.ServiceIsUnavailable
	res.Message = packed.Err.GetErrorMessage(consts.ServiceIsUnavailable)
	res.Data = nil
	// 验证 req.Host 是否存于数据库中
	//valCheckUrl := service.User().IsDomainExist(ctx, model.CheckUrlInput{Host: req.Host})
	//if !valCheckUrl {
	//res.Message = fmt.Sprintf(`"%s"`, gerror.Newf(`Sorry, your domain name "%s" is not registered yet`, req.Host))
	//return res, nil
	//}
	client := hub.GetClient(req.ClientID)
	if client == nil {
		return res, nil
	}
	clientIpc := client.GetIpc()
	reqIpc := clientIpc.Request(req.URL, ipc.RequestArgs{
		Method: req.Method,
		Header: map[string]string{"Content-Type": req.Header},
	})
	resIpc, err := clientIpc.Send(ctx, reqIpc)
	if err != nil {
		log.Println("ipc response err: ", err)
		res.Code = consts.ClientIpcSendErr
		res.Message = err.Error()
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
		res.Message = err.Error()
		return res, err
	}
	res.Code = consts.Success
	res.Message = packed.Err.GetErrorMessage(consts.Success)
	res.Data = string(resStr)
	return res, err
}
