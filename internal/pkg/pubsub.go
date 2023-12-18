package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BioforestChain/dweb-browser-proxy/internal/consts"
	redisHelper "github.com/BioforestChain/dweb-browser-proxy/internal/pkg/redis"
	"github.com/BioforestChain/dweb-browser-proxy/internal/pkg/ws"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/ipc"
	"github.com/redis/go-redis/v9"
	"io"
	"log"
	"net/url"
)

type Cache struct {
	Ctx      context.Context
	RedisCli *redisHelper.RedisInstance
}

func NewCache(ctx context.Context) *Cache {
	redisCli, _ := redisHelper.GetRedisInstance("default")
	return &Cache{Ctx: ctx, RedisCli: redisCli}
}

type IpcBodyData struct {
	Topic string `json:"topic"`
	Data  string `json:"data"`
}
type IpcHeaderData struct {
	XDwebHost string `json:"X-Dweb-Host"` // X-Dweb-Pubsub
	//XDWebPubSub       string `json:"X-Dweb-Pubsub"`
	XDWebPubSubApp string `json:"X-Dweb-Pubsub-App"`
	//XDWebPubSubNet    string `json:"X-Dweb-Pubsub-Net"`
	//XDWebPubSubDomain string `json:"X-Dweb-Pubsub-Net-Domain"`
}

// IpcEvent.data
type IpcEventDataHeaderBody struct {
	Headers IpcHeaderData `json:"headers"`
	Body    IpcBodyData   `json:"body"`
}

func getCacheKey(keyName string) string {
	return fmt.Sprintf(consts.FormatKey, consts.RedisPrefix, keyName)
}

//var clientIPC = ipc.NewReadableStreamIPC(ipc.CLIENT, ipc.SupportProtocol{
//	Raw:         true,
//	MessagePack: false,
//	ProtoBuf:    false,
//})

//func TestClientIPCOnRequest(ctx context.Context, hub *Hub, w http.ResponseWriter, r *http.Request) {
//
//	client := &Client{
//		ID:  r.URL.Query().Get("client_id"), // TODO 用户id
//		hub: hub,
//		//conn: conn,
//		// //send:        make(chan []byte, 256),
//		ipc:         clientIPC,
//		inputStream: ipc.NewReadableStream(),
//	}
//	//TODO 模拟数据
//	// ~~~~~~
//	header := map[string]string{
//		"X-Dweb-Host":              "mmid2",
//		"X-Dweb-Pubsub":            "mmid2",
//		"X-Dweb-Pubsub-App":        "app_mmid2",
//		"X-Dweb-Pubsub-Net":        "net_mmid2",
//		"X-Dweb-Pubsub-Net-Domain": "userName_domain3",
//	}
//
//	ipcBody := map[string]string{"topic": "topic_name_xx"}
//	bodyData, err := json.Marshal(ipcBody)
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	bodySubSender := ipc.NewBodySender(bodyData, clientIPC)
//	// ~~~~~~
//	//TODO 模拟数据 发起IPC request
//	//对接 js模块
//	//go func() {
//	clientIPC.MsgSignal.Emit(ipc.NewRequest(1, "/sub", "POST", header, bodySubSender, clientIPC), nil)
//	//}()
//
//	clientIPC.OnRequest(func(data any, ipcObj ipc.IPC) {
//		client := hub.GetClient(client.ID)
//		if client == nil {
//			return
//		}
//		request := data.(*ipc.Request)
//		if request.URL == "/sub" && request.Method == ipc.POST {
//			handlerSub(ctx, request, client)
//		}
//	})
//}
//
//func TestClientIPCOnRequestPub(ctx context.Context, hub *Hub, w http.ResponseWriter, r *http.Request) {
//	client := &Client{
//		ID:  r.URL.Query().Get("client_id"), // TODO 用户id
//		hub: hub,
//		//conn: conn,
//		// //send:        make(chan []byte, 256),
//		ipc:         clientIPC,
//		inputStream: ipc.NewReadableStream(),
//	}
//	//TODO 模拟数据
//	// ~~~~~~
//	header := map[string]string{
//		"X-Dweb-Host":              "mmid2",
//		"X-Dweb-Pubsub":            "mmid2",
//		"X-Dweb-Pubsub-App":        "app_mmid2",
//		"X-Dweb-Pubsub-Net":        "net_mmid2",
//		"X-Dweb-Pubsub-Net-Domain": "userName_domain3",
//	}
//	// ~~~~~~
//	ipcBodyPub := map[string]string{
//		"topic": "topic_name_xx",
//		"data":  "{\"success\":false,\"message\":\"Not Found\"}",
//	}
//	strPub, err := json.Marshal(ipcBodyPub)
//	if err != nil {
//		fmt.Println(err)
//	}
//	bodyPubSender := ipc.NewBodySender(strPub, clientIPC)
//
//	go func() {
//		clientIPC.MsgSignal.Emit(ipc.NewRequest(1, "/pub", "POST", header, bodyPubSender, clientIPC), nil)
//	}()
//
//	clientIPC.OnRequest(func(data any, ipcObj ipc.IPC) {
//		client := hub.GetClient(client.ID)
//		if client == nil {
//			return
//		}
//		request := data.(*ipc.Request)
//		if request.URL == "/pub" && request.Method == ipc.POST {
//			handlerPub(ctx, request)
//		}
//	})
//}

var DefaultPubSub = NewPubSub()

type PubSub struct {
}

func NewPubSub() *PubSub {
	return &PubSub{}
}

func (pb *PubSub) Handler(ctx context.Context, request *ipc.Request, client *ws.Client) (err error) {
	parsedURL, err := url.Parse(request.URL)
	if err != nil {
		return
	}

	subPath := fmt.Sprintf("/%s/sub", request.Header.Get("X-Dweb-Pubsub-App"))
	if parsedURL.Path == subPath && request.Method == ipc.POST {
		if err = pb.Sub(ctx, request, client); err != nil {
			return
		}
	}

	pubPath := fmt.Sprintf("/%s/pub", request.Header.Get("X-Dweb-Pubsub-App"))
	if parsedURL.Path == pubPath && request.Method == ipc.POST {
		if err = pb.Pub(ctx, request); err != nil {
			return
		}
	}

	return
}

// Sub
// 处理Sub逻辑：订阅请求,生成topic和net domain对应关系
//
//	@Description:
//	@param ctx
//	@param request
//	@param ipcBodyData
//	@param client
//	@return err
func (pb *PubSub) Sub(ctx context.Context, request *ipc.Request, client *ws.Client) (err error) {
	var ipcBodyData IpcBodyData

	//body := make([]byte, 0)
	//bodyStream := request.Body.Stream()
	//if bodyStream != nil {
	//	if body, err = helperIPC.ReadStreamWithTimeout(bodyStream, 5*time.Second); err != nil {
	//		return
	//	}
	//}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		return
	}

	if err = json.Unmarshal(body, &ipcBodyData); err != nil {
		// TODO 日志上报
		return
	}

	getXDWebPubSub := request.Header["X-Dweb-Pubsub"]
	//getXDWebHost := request.Header["X-Dweb-Host"]
	getXDWebPubSubDomain := request.Header["X-Dweb-Pubsub-Net-Domain"]
	getXDWebPubSubApp := request.Header["X-Dweb-Pubsub-App"]
	//getXDWebPubSubNet := request.Header["X-Dweb-Pubsub-Net"]
	getTopicName := ipcBodyData.Topic

	// 处理重复订阅
	// 服务端重启时会导致所有订阅goroutine退出，所以判断是否重复订阅应该基于订阅goroutine是否存活？
	exist, err := NewCache(ctx).RedisCli.SIsMember(ctx, getCacheKey(getTopicName), getXDWebPubSubDomain)
	if (exist || err != nil) && client.Online() {
		return
	}

	// 存储映射
	// topic ----netDomain
	_, err = NewCache(ctx).RedisCli.SAdd(ctx, getCacheKey(getTopicName), getXDWebPubSubDomain)
	if err != nil {
		return
	}

	ctxChild, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-client.Shutdown:
			fmt.Println("channel closed")
			cancel()
		}
	}()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("============panic Sub callback getPub's data ============", err)
			}
		}()
		err = NewCache(ctxChild).RedisCli.Sub(ctxChild, func(data *redis.Message) error {
			//分发
			netDomains, err := NewCache(ctx).RedisCli.SMembers(ctx, getCacheKey(getTopicName))
			for _, netDomain := range netDomains {
				go func(netDomain string) {
					defer func() {
						if err := recover(); err != nil {
							log.Println("go handlerSub panic: ", err)
						}
					}()
					//IpcEvent.data = {
					//headers: {
					//	"X-Dweb-Host": "xxx.dweb"， // 必填，网络模块转发的下个模块mmid；发布订阅模式下，就是发布订阅模块mmid
					//	"X-Dweb-Pubsub-App": "xxx.dweb", // 选填，发布订阅模式下，是使用发布订阅模块的mmid
					//},
					//body: {
					//topic: "xxx" // 必填，订阅的主题
					//	data：string || []byte ， // 必填，发布的数据   //data.Payload
					//}
					//}
					ipcCombHeaderBody := new(IpcEventDataHeaderBody)
					ipcCombHeaderBody.Headers = IpcHeaderData{
						XDwebHost:      getXDWebPubSub,
						XDWebPubSubApp: getXDWebPubSubApp,
					}
					ipcCombHeaderBody.Body = IpcBodyData{
						getTopicName, data.Payload,
					}
					bodyData, err := json.Marshal(ipcCombHeaderBody)
					if err != nil {
						log.Println("json ipcEventDataHeaderBody err is: ", err)
						return
					}
					eventData := ipc.NewEventBase64(getTopicName, bodyData)
					fmt.Printf("eventData data is :%#v\n", eventData)

					netClient := client.GetHub().GetClient(netDomain)
					if netClient == nil {
						log.Println("not found client: ", netDomain)
						return
					}

					if err := netClient.GetIpc().PostMessage(ctx, eventData); err != nil {
						log.Println("ipc PostMessage err is: ", err)
						return
					}

				}(netDomain)
			}

			if err != nil {
				log.Println("RedisCli SMembers panic: ", err)
				return err
			}
			return nil
		}, getTopicName)

		if err != nil {
			// TODO
			log.Println("RedisCli sub err: ", err)
		}
	}()

	return
}

// Pub
//
//	@Description: 处理Pub逻辑
//	@param ctx
//	@param request
//	@param ipcBodyData
//	@param client
//	@return err
func (pb *PubSub) Pub(ctx context.Context, request *ipc.Request) (err error) {
	var ipcBodyData IpcBodyData

	body, err := io.ReadAll(request.Body)
	if err != nil {
		return
	}

	if err = json.Unmarshal(body, &ipcBodyData); err != nil {
		return
	}

	getTopicName := ipcBodyData.Topic
	getTopicData := ipcBodyData.Data

	// TODO 发布数据时，如果用户未订阅，则不进行订阅操作，同时返回订阅提醒
	_, err = NewCache(ctx).RedisCli.Pub(ctx, getTopicName, getTopicData)
	if err != nil {
		return
	}

	return nil
}
