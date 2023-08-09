package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"proxyServer/ipc"
)

func main() {
	url := "ws://127.0.0.1:8000/ws"
	conn, res, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalln("Error connecting to server: ", err)
	}
	defer conn.Close()

	log.Println("Connected to WebSocket server", res.Status)

	ipcConn := newIPCConn(conn)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message: ", err)
			return
		}
		fmt.Printf("Received message: %s\n", message)

		ipcConn.proxyStream.Controller.Enqueue(message)
	}

}

type IPCConn struct {
	ipc         ipc.IPC
	proxyStream *ipc.ReadableStream
	conn        *websocket.Conn
}

func newIPCConn(conn *websocket.Conn) *IPCConn {
	serverIPC := ipc.NewReadableStreamIPC(ipc.SERVER, ipc.SupportProtocol{
		Raw:         true,
		MessagePack: false,
		ProtoBuf:    false,
	})

	// 监听并处理请求，echo请求数据
	serverIPC.OnRequest(func(req interface{}, ic ipc.IPC) {
		request := req.(*ipc.Request)
		log.Println("on request: ", request.URL)

		url, _ := url.ParseRequestURI(request.URL)

		if (url.Host + url.Path) == "www.example.com/search" {
			bodyReceiver := request.Body.(*ipc.BodyReceiver)
			body := bodyReceiver.GetMetaBody().Data
			log.Println("onRequest: ", request.URL, string(body), ic)

			res := ipc.NewResponse(
				request.ID,
				200,
				ipc.NewHeaderWithExtra(map[string]string{
					"Content-Type": "application/json",
				}),
				ipc.NewBodySender([]byte("hi"), ic),
				ic,
			)

			if err := ic.PostMessage(res); err != nil {
				log.Println("post message err: ", err)
			}
		}
	})

	proxyStream := ipc.NewReadableStream()

	ipcConn := &IPCConn{
		ipc:         serverIPC,
		proxyStream: proxyStream,
		conn:        conn,
	}

	go func() {
		defer ipcConn.Close()
		// 读取proxyStream数据并emit消息（接收消息并处理，然后把结果发送至内部流）
		if err := serverIPC.BindIncomeStream(proxyStream); err != nil {
			panic(err)
		}
	}()

	go func() {
		// 读取内部流数据，然后response
		serverIPC.ReadFromStream(func(data []byte) {
			log.Println("write msg: ", string(data))
			if err := ipcConn.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				panic(err)
			}
		})
	}()

	return ipcConn
}

func (i *IPCConn) Close() {
	i.ipc.Close()
	i.proxyStream.Controller.Close()
	i.conn.Close()
}
