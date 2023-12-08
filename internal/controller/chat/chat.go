package chat

import (
	"context"
	"fmt"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/redis/go-redis/v9"
	v1 "proxyServer/api/chat/v1"
	v1Client "proxyServer/api/client/v1"
	redisHelper "proxyServer/internal/helper/redis"
	ws "proxyServer/internal/service/ws"
)

type Controller struct {
	hub *ws.Hub
}

func New(hub *ws.Hub) *Controller {
	return &Controller{
		hub: hub,
	}
}

type Cache struct {
	Ctx      context.Context
	RedisCli *redisHelper.RedisInstance
}

func NewCache(ctx context.Context) *Cache {
	redisCli, _ := redisHelper.GetRedisInstance("default")
	return &Cache{Ctx: ctx, RedisCli: redisCli}
}

// CreateTopicReq
//
//	@Description:
//	@receiver c
//	@param ctx
//	@param req
//	@return res
//	@return err
func (c *Controller) CreateTopicReq(ctx context.Context, req *v1.CreateTopicReq) (res *v1.CreateTopicRes, err error) {
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("ChatCreateTopicReq Validator", err)
	}
	_, err = NewCache(ctx).RedisCli.Pub(ctx, req.TopicName, req.Msg)
	//_, err = g.Redis().Publish(ctx, req.TopicName, req.Msg)
	if err != nil {
		g.Log().Debug(ctx, err)
		return nil, err
	}
	return
}

// SubscribeMsgReq /proxy/pubsub/subscribe_msg
//
//	@Description:
//	@receiver c
//	@param ctx
//	@param req
//	@return res
//	@return err
func (c *Controller) SubscribeMsgReq(ctx context.Context, req *v1.SubscribeMsgReq) (res *v1.SubscribeMsgRes, err error) {
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("SubscribeMsgReq Validator", err)
	}

	//todo 离线--hub信号--监听关闭协程
	gRequestData := g.RequestFromCtx(ctx)

	var ctxChild = context.Background()
	go func() {
		select {
		case <-c.hub.Shutdown:
			ctxChild.Done()
		}
	}()

	go func() {
		err = NewCache(ctxChild).RedisCli.Sub(ctxChild, func(data *redis.Message) error {
			reqC := new(v1Client.IpcReq)
			reqC.Header = gRequestData.Header
			reqC.Method = gRequestData.Method
			reqC.URL = gRequestData.GetUrl()
			reqC.Host = gRequestData.GetHost()
			reqC.Body = gRequestData.GetBody()
			reqC.ClientID = gRequestData.Get("client_id").String()
			reqC.Body = data.Payload
			_, err = ws.Proxy2Ipc(ctxChild, c.hub, reqC)
			return err
		}, req.TopicName)

		//conn, _, err := g.Redis().Subscribe(ctx, req.TopicName)

		if err != nil {
			g.Log().Debug(ctxChild, err)
			//return nil, err
		}
	}()
	return res, nil
}
