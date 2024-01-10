package controller

import (
	"context"
	"errors"
	"fmt"
	v1Client "github.com/BioforestChain/dweb-browser-proxy/api/client/v1"
	pubsub2 "github.com/BioforestChain/dweb-browser-proxy/app/pubsub"
	"github.com/BioforestChain/dweb-browser-proxy/app/pubsub/api/chat/v1"
	"github.com/BioforestChain/dweb-browser-proxy/app/pubsub/consts"
	redisHelper "github.com/BioforestChain/dweb-browser-proxy/pkg/redis"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/ws"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/redis/go-redis/v9"
)

type chat struct {
	hub *ws.Hub
}

func NewChat(hub *ws.Hub) *chat {
	return &chat{
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
func (c *chat) CreateTopicReq(ctx context.Context, req *v1.CreateTopicReq) (res *v1.CreateTopicRes, err error) {
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
//
// 离线--client 信号--监听关闭协程--防止无效订阅协程常存
//
//	@receiver c
//	@param ctx
//	@param req
//	@return res
//	@return err

func (c *chat) SubscribeMsgReq(ctx context.Context, req *v1.SubscribeMsgReq) (res *v1.SubscribeMsgRes, err error) {
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("SubscribeMsgReq Validator", err)
		return nil, err
	}
	gRequestData := g.RequestFromCtx(ctx)
	topicName := "testmodule.bagen.com.dweb_topic_112233"
	xDWebHostDomain := gRequestData.Header.Get(consts.XDwebHostDomain)

	pubsub2.CheckExistNetDomainInAclList(ctx, topicName, xDWebHostDomain)

	ctxChild, cancel := context.WithCancel(context.Background())
	clientId := gRequestData.Get("client_id").String()
	client := c.hub.GetClient(clientId)
	if client == nil {
		return
	}

	go func() {
		select {
		case <-client.Shutdown:
			fmt.Println("channel closed")
			//ctxChild.Done()
			cancel()
		default:
		}
	}()
	//第二个协程
	go func() {
		err = NewCache(ctxChild).RedisCli.Sub(ctxChild, func(data *redis.Message) error {
			reqC := new(v1Client.IpcReq)
			reqC.Header = gRequestData.Header
			reqC.Method = gRequestData.Method
			reqC.URL = gRequestData.GetUrl()
			reqC.Host = gRequestData.GetHost()
			reqC.Body = gRequestData.GetBody()
			reqC.ClientID = clientId
			reqC.Body = data.Payload

			client := c.hub.GetClient(clientId)
			if client == nil {
				return errors.New("the service is unavailable~")
			}
			_, err = ws.SendIPC(ctxChild, client, reqC)
			fmt.Println("ctxSrcRelease Proxy2Ipc", err)
			return err
		}, req.TopicName)

		if err != nil {
			fmt.Println("ctxSrcRelease Validator", err)
			//return
		}
		//第一个协程
		//go func() {
		//	client := c.hub.GetClient(clientId)
		//	if client == nil {
		//		return
		//	}
		//	c.hub.EndSyncCond.L.Lock()
		//	for !client.DisConn {
		//		c.hub.EndSyncCond.Wait()
		//	}
		//	cancel()
		//	c.hub.EndSyncCond.L.Unlock()
		//}()

	}()
	return res, nil
}
