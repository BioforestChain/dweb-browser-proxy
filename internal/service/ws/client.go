package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"net/url"
	v1Client "proxyServer/api/client/v1"
	"proxyServer/internal/consts"
	helperIPC "proxyServer/internal/helper/ipc"
	redisHelper "proxyServer/internal/helper/redis"
	"proxyServer/ipc"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	//writeWait = 10 * time.Second
	writeWait = 1000 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 600 * time.Second
	//pongWait = 7 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	//pingPeriod = 4 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 8
)

type Cache struct {
	Ctx      context.Context
	RedisCli *redisHelper.RedisInstance
}

func NewCache(ctx context.Context) *Cache {
	redisCli, _ := redisHelper.GetRedisInstance("default")
	return &Cache{Ctx: ctx, RedisCli: redisCli}
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	ID string

	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound message.
	//send chan []byte

	ipc ipc.IPC

	inputStream *ipc.ReadableStream

	closed bool

	mutex sync.Mutex
}

// readPump pumps messages from the websocket connection to the hub.
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from the goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister <- c
		//_ = c.conn.Close()
		c.Close()

		if err := recover(); err != nil {
			// TODO 日志上报
			log.Println("readPump ws panic: ", err)
		}
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(s string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("readPump err: ", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		if err := c.inputStream.Enqueue(message); err != nil {
			log.Println("inputStream Enqueue err: ", err)
		}

		//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		//c.hub.broadcast <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	reader := c.ipc.GetOutputStreamReader()

	defer func() {
		ticker.Stop()
		//_ = c.conn.Close()
		c.Close()
		reader.Cancel()

		if err := recover(); err != nil {
			// TODO 日志上报
			log.Println("writePump ws panic: ", err)
		}
	}()

	go func() {
		for {
			<-ticker.C
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				reader.Cancel()
				return
			}
		}
	}()

	for {
		message, err := reader.Read()
		_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err != nil || message.Done {
			// The hub closed the channel.
			_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		w, err := c.conn.NextWriter(websocket.BinaryMessage)
		if err != nil {
			return
		}

		_, _ = w.Write(message.Value)

		// Add queued messages to the current websocket message.
		//n := len(c.send)
		//for i := 0; i < n; i++ {
		//	_, _ = w.Write(newline)
		//	_, _ = w.Write(<-c.send)
		//}

		if err := w.Close(); err != nil {
			log.Println("writePump close: ", err)
			return
		}
	}
}

func (c *Client) GetIpc() ipc.IPC {
	return c.ipc
}

func (c *Client) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return
	}
	c.closed = true

	c.ipc.Close()
	c.inputStream.Controller.Close()
	_ = c.conn.Close()
}

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		// Origin header have a pattern that *.xxx.com
		// TODO return r.Header.Get("Origin") == '*.xxx.com'
		return true
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	clientIPC := ipc.NewReadableStreamIPC(ipc.CLIENT, ipc.SupportProtocol{
		Raw:         true,
		MessagePack: false,
		ProtoBuf:    false,
	})

	client := &Client{
		ID:   r.URL.Query().Get("client_id"), // TODO 用户id
		hub:  hub,
		conn: conn,
		//send:        make(chan []byte, 256),
		ipc:         clientIPC,
		inputStream: ipc.NewReadableStream(),
	}

	client.hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines.
	go client.writePump()
	go client.readPump()

	go func() {
		defer func() {
			client.Close()
			if err := recover(); err != nil {
				// TODO 日志上报
				log.Println("clientIPC.BindInputStream panic: ", err)
			}
		}()

		if err := clientIPC.BindInputStream(client.inputStream); err != nil {
			log.Println("clientIPC.BindInputStream: ", err)
		}
	}()

	clientIPC.OnRequest(func(data any, ipcObj ipc.IPC) {
		request := data.(*ipc.Request)

		if len(request.Header.Get("X-Dweb-Pubsub")) > 0 {
			if err := handlerPubSub(request, client); err != nil {
				// TODO
				log.Println("handlerPubSub err: ", err)
				body := []byte(fmt.Sprintf(`{"success": false, "message": "%s"}`, err.Error()))
				err = clientIPC.PostMessage(context.Background(), ipc.FromResponseBinary(request.ID, http.StatusOK, ipc.NewHeader(), body, ipcObj))
				fmt.Println("PostMessage err: ", err)
				return
			}

			body := []byte(fmt.Sprint(`{"success": true, "message": "ok"}`))
			err = clientIPC.PostMessage(context.Background(), ipc.FromResponseBinary(request.ID, http.StatusOK, ipc.NewHeader(), body, ipcObj))
			fmt.Println("PostMessage err: ", err)
		}
	})
}

type IpcBodyData struct {
	Topic string `json:"topic"`
	Data  string `json:"data"`
}
type IpcHeaderData struct {
	XDWebPubSub       string `json:"X-Dweb-Pubsub"`
	XDWebPubSubApp    string `json:"X-Dweb-Pubsub-App"`
	XDWebPubSubNet    string `json:"X-Dweb-Pubsub-Net"`
	XDWebPubSubDomain string `json:"X-Dweb-Pubsub-Net-Domain"`
}

func getCacheKey(keyName string) string {
	return fmt.Sprintf(consts.FormatKey, consts.RedisPrefix, keyName)
}

var clientIPC = ipc.NewReadableStreamIPC(ipc.CLIENT, ipc.SupportProtocol{
	Raw:         true,
	MessagePack: false,
	ProtoBuf:    false,
})

func ClientIPCOnRequest(ctx context.Context, hub *Hub, w http.ResponseWriter, r *http.Request) {

	client := &Client{
		ID:  r.URL.Query().Get("client_id"), // TODO 用户id
		hub: hub,
		//conn: conn,
		// //send:        make(chan []byte, 256),
		ipc:         clientIPC,
		inputStream: ipc.NewReadableStream(),
	}
	//TODO 模拟数据
	// ~~~~~~
	header := map[string]string{
		"X-Dweb-Pubsub":            "mmid2",
		"X-Dweb-Pubsub-App":        "app_mmid2",
		"X-Dweb-Pubsub-Net":        "net_mmid2",
		"X-Dweb-Pubsub-Net-Domain": "userName_domain3",
	}

	ipcBody := map[string]string{"topic": "topic_name_xx"}
	str, err := json.Marshal(ipcBody)
	if err != nil {
		fmt.Println(err)
	}

	bodySubSender := ipc.NewBodySender(str, clientIPC)
	// ~~~~~~
	//TODO 模拟数据 发起IPC request
	//对接 js模块
	go func() {
		clientIPC.MsgSignal.Emit(ipc.NewRequest(1, "/sub", "POST", header, bodySubSender, clientIPC), nil)
	}()

	clientIPC.OnRequest(func(data any, ipcObj ipc.IPC) {
		request := data.(*ipc.Request)
		if request.URL == "/sub" && request.Method == ipc.POST {
			handlerSub(ctx, request, client)
		}
	})
}

func ClientIPCOnRequestPub(ctx context.Context, hub *Hub, w http.ResponseWriter, r *http.Request) {

	var ipcBodyData IpcBodyData

	client := &Client{
		ID:  r.URL.Query().Get("client_id"), // TODO 用户id
		hub: hub,
		//conn: conn,
		// //send:        make(chan []byte, 256),
		ipc:         clientIPC,
		inputStream: ipc.NewReadableStream(),
	}
	//TODO 模拟数据
	// ~~~~~~
	header := map[string]string{
		"X-Dweb-Pubsub":            "mmid2",
		"X-Dweb-Pubsub-App":        "app_mmid2",
		"X-Dweb-Pubsub-Net":        "net_mmid2",
		"X-Dweb-Pubsub-Net-Domain": "userName_domain3",
	}
	// ~~~~~~
	ipcBodyPub := map[string]string{
		"topic": "topic_name_xx",
		"data":  "{\"success\":false,\"message\":\"Not Found\"}",
	}
	strPub, err := json.Marshal(ipcBodyPub)
	if err != nil {
		fmt.Println(err)
	}
	bodyPubSender := ipc.NewBodySender(strPub, clientIPC)

	go func() {
		clientIPC.MsgSignal.Emit(ipc.NewRequest(1, "/pub", "POST", header, bodyPubSender, clientIPC), nil)
	}()

	clientIPC.OnRequest(func(data any, ipcObj ipc.IPC) {
		request := data.(*ipc.Request)
		if request.URL == "/pub" && request.Method == ipc.POST {
			handlerPub(ctx, request, ipcBodyData, client)
		}
	})
}

func handlerPubSub(request *ipc.Request, client *Client) error {
	parsedURL, err := url.Parse(request.URL)
	if err != nil {
		return err
	}

	subPath := fmt.Sprintf("/%s/sub", request.Header.Get("X-Dweb-Pubsub-App"))
	if parsedURL.Path == subPath && request.Method == ipc.POST {
		if err := handlerSub(context.Background(), request, client); err != nil {
			// TODO
			return err
		}
	}

	pubPath := fmt.Sprintf("/%s/pub", request.Header.Get("X-Dweb-Pubsub-App"))
	if request.URL == pubPath && request.Method == ipc.POST {

	}

	return nil
}

// handlerSub
// 处理Sub逻辑：订阅请求,生成topic和net domain对应关系
//
//	@Description:
//	@param ctx
//	@param request
//	@param ipcBodyData
//	@param client
//	@return err
func handlerSub(ctx context.Context, request *ipc.Request, client *Client) (err error) {
	fmt.Printf("Header:%#v\n", request.Header)
	var ipcBodyData IpcBodyData

	body := make([]byte, 0)
	bodyStream := request.Body.Stream()
	if bodyStream != nil {
		if body, err = helperIPC.ReadStreamWithTimeout(bodyStream, 5*time.Second); err != nil {
			return err
		}
	}

	if err = json.Unmarshal(body, &ipcBodyData); err != nil {
		// TODO 日志上报
		return err
	}

	getXDWebPubSub := request.Header["X-Dweb-Pubsub"]
	getXDWebPubSubDomain := request.Header["X-Dweb-Pubsub-Net-Domain"]
	getXDWebPubSubApp := request.Header["X-Dweb-Pubsub-App"]
	getXDWebPubSubNet := request.Header["X-Dweb-Pubsub-Net"]
	getTopicName := ipcBodyData.Topic

	fmt.Printf("getXDWebPubSub:%#v\n", getXDWebPubSub)
	fmt.Printf("getXDWebPubSubNet:%#v\n", getXDWebPubSubNet)

	//存储映射
	//topic ----netDomain
	_, err = NewCache(ctx).RedisCli.SAdd(ctx, getCacheKey(getTopicName), getXDWebPubSubDomain)

	if err != nil {
		return err
	}

	//发起订阅
	var ctxChild = context.Background()
	go func() {
		select {
		case <-client.hub.Shutdown:
			ctxChild.Done()
		}
	}()

	headerData := map[string][]string{
		"X-Dweb-Pubsub":            {getXDWebPubSub},
		"X-Dweb-Pubsub-App":        {getXDWebPubSubApp},
		"X-Dweb-Pubsub-Net":        {getXDWebPubSubNet},
		"X-Dweb-Pubsub-Net-Domain": {getXDWebPubSubDomain},
	}

	err = NewCache(ctxChild).RedisCli.Sub(ctxChild, func(data *redis.Message) error {
		reqC := new(v1Client.IpcReq)
		//分发
		userList, err := NewCache(ctx).RedisCli.SMembers(ctx, getCacheKey(getTopicName))
		for _, usr := range userList {
			reqC.Method = string(request.Method)
			reqC.URL = request.URL
			reqC.Body = body
			reqC.ClientID = usr
			reqC.Body = data.Payload
			headerData["X-Dweb-Pubsub-Net-Domain"][0] = usr
			reqC.Header = headerData
			go func() {
				defer func() {
					if err := recover(); err != nil {
						log.Println("go handlerSub panic: ", err)
					}
				}()
				fmt.Printf("reqC:%#v\n", reqC)
				response, err := Proxy2Ipc(ctxChild, client.hub, reqC)
				fmt.Printf("resPonse data is :%#v\n", response)
				if err != nil {
					log.Println("RedisCli Sub panic: ", err)
				}
			}()
		}
		if err != nil {
			log.Println("RedisCli SMembers panic: ", err)
			return err
		}
		return nil
	}, getTopicName)

	//go func() {
	//	defer func() {
	//		if err := recover(); err != nil {
	//			fmt.Println("============panic Sub callback getPub's data ============", err)
	//		}
	//	}()
	//	headerData := map[string][]string{
	//		"X-Dweb-Pubsub":            {getXDWebPubSub},
	//		"X-Dweb-Pubsub-App":        {getXDWebPubSubApp},
	//		"X-Dweb-Pubsub-Net":        {getXDWebPubSubNet},
	//		"X-Dweb-Pubsub-Net-Domain": {getXDWebPubSubDomain},
	//	}
	//	err = NewCache(ctxChild).RedisCli.Sub(ctxChild, func(data *redis.Message) error {
	//		reqC := new(v1Client.IpcReq)
	//		//分发
	//		userList, err := NewCache(ctx).RedisCli.SMembers(ctx, getCacheKey(getTopicName))
	//		for _, usr := range userList {
	//			reqC.Method = string(request.Method)
	//			reqC.URL = request.URL
	//			reqC.Body = body
	//			reqC.ClientID = usr
	//			reqC.Body = data.Payload
	//			headerData["X-Dweb-Pubsub-Net-Domain"][0] = usr
	//			reqC.Header = headerData
	//			go func() {
	//				defer func() {
	//					if err := recover(); err != nil {
	//						log.Println("go handlerSub panic: ", err)
	//					}
	//				}()
	//				fmt.Printf("reqC:%#v\n", reqC)
	//				response, err := Proxy2Ipc(ctxChild, client.hub, reqC)
	//				fmt.Printf("resPonse data is :%#v\n", response)
	//				if err != nil {
	//					log.Println("RedisCli Sub panic: ", err)
	//				}
	//			}()
	//		}
	//		if err != nil {
	//			log.Println("RedisCli SMembers panic: ", err)
	//		}
	//		return nil
	//	}, getTopicName)
	//	if err != nil {
	//		log.Println("RedisCli Sub panic: ", err)
	//	}
	//}()

	return err
}

// handlerPub
//
//	@Description: 处理Pub逻辑
//	@param ctx
//	@param request
//	@param ipcBodyData
//	@param client
//	@return err
func handlerPub(ctx context.Context, request *ipc.Request, ipcBodyData IpcBodyData, client *Client) (err error) {
	fmt.Printf("Header:%#v\n", request.Header)
	body := request.Body.GetMetaBody().Data
	json.Unmarshal(body, &ipcBodyData)
	getTopicName := ipcBodyData.Topic
	getTopicData := ipcBodyData.Data

	//发起发布消息
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("============panic handlerPub============", err)
			}
		}()
		_, err = NewCache(ctx).RedisCli.Pub(ctx, getTopicName, getTopicData)
		if err != nil {
			log.Println("RedisCli Pub panic: ", err)
		}
	}()
	return nil
}

// Proxy2Ipc
//
//	@Description: The request goes to the IPC object for processing
//	@param ctx
//	@param hub
//	@param req
//	@return res
//	@return err
func Proxy2Ipc(ctx context.Context, hub *Hub, req *v1Client.IpcReq) (res *ipc.Response, err error) {
	client := hub.GetClient(req.ClientID)
	if client == nil {
		return nil, errors.New("the service is unavailable~")
	}
	var (
		clientIpc     = client.GetIpc()
		overallHeader = make(map[string]string)
	)
	for k, v := range req.Header {
		overallHeader[k] = v[0]
	}
	reqIpc := clientIpc.Request(req.URL, ipc.RequestArgs{
		Method: req.Method,
		Header: overallHeader,
		Body:   req.Body,
	})
	resIpc, err := clientIpc.Send(ctx, reqIpc)
	if err != nil {
		return nil, err
	}
	return resIpc, nil
}
