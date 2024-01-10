package middleware

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"golang.org/x/time/rate"
	"net/http"
	"sync"
	"time"
)

func LimitHandler() func(r *ghttp.Request) {
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
			r.Response.Write(Response{http.StatusTooManyRequests, "The request is too fast, please try again later!", nil})
		}
	}
}
