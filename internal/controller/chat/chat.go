package chat

import (
	"context"
	"fmt"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/frame/g"
	"io"
	v1 "proxyServer/api/chat/v1"
	v1Client "proxyServer/api/client/v1"
	"proxyServer/internal/consts"
	helperIPC "proxyServer/internal/helper/ipc"
	"proxyServer/internal/packed"
	ws "proxyServer/internal/service/ws"
	"time"
)

type Controller struct {
	hub *ws.Hub
}

func New(hub *ws.Hub) *Controller {
	return &Controller{
		hub: hub,
	}
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
	_, err = g.Redis().Publish(ctx, req.TopicName, req.Msg)
	if err != nil {
		g.Log().Debug(ctx, err)
		return nil, err
	}
	return
}

// SubscribeMsgReq
//
//	@Description: /proxy/pubsub/subscribe_msg
//	@receiver c
//	@param ctx
//	@param req
//	@return res
//	@return err
func (c *Controller) SubscribeMsgReq(ctx context.Context, req *v1.SubscribeMsgReq) (res *v1.SubscribeMsgRes, err error) {
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("SubscribeMsgReq Validator", err)
	}
	conn, _, err := g.Redis().Subscribe(ctx, req.TopicName)
	if err != nil {
		g.Log().Debug(ctx, err)
		return nil, err
	}
	reqC := new(v1Client.IpcReq)
	gRequestData := g.RequestFromCtx(ctx)
	reqC.Header = gRequestData.Header
	reqC.Method = gRequestData.Method
	reqC.URL = gRequestData.GetUrl()
	reqC.Host = gRequestData.GetHost()
	reqC.Body = gRequestData.GetBody()
	reqC.ClientID = gRequestData.Get("client_id").String()
	go func() {
		var ctx = context.Background()
		for {
			msg, _ := conn.ReceiveMessage(ctx)
			time.Sleep(1 * time.Second)
			fmt.Printf("SubscribeMsg.Payload:%#v\n", msg.Payload)
			resIpc, err := packed.Proxy2Ipc(ctx, c.hub, reqC)
			if err != nil {
				resIpc = packed.IpcErrResponse(consts.ServiceIsUnavailable, err.Error())
			}
			for k, v := range resIpc.Header {
				gRequestData.Response.Header().Set(k, v)
			}
			bodyStream := resIpc.Body.Stream()
			if bodyStream == nil {
				if _, err = io.Copy(gRequestData.Response.Writer, resIpc.Body); err != nil {
					gRequestData.Response.WriteStatus(400, "请求出错")
				}
			} else {
				data, err := helperIPC.ReadStreamWithTimeout(bodyStream, 10*time.Second)
				if err != nil {
					gRequestData.Response.WriteStatus(400, err)
				} else {
					gRequestData.Response.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
					_, _ = gRequestData.Response.Writer.Write(data)
				}
			}
		}
	}()
	return res, nil
}
