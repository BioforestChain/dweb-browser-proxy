package main

import (
	"context"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"proxyServer/ipc"
)

func main() {
	url := "ws://127.0.0.1:8000/ws?client_id=8cb46dde8d8edb41994e0b88f87a31dc"
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
		//fmt.Printf("Received message: %s\n", message)

		if err := ipcConn.inputStream.Enqueue(message); err != nil {
			log.Println("enqueue err: ", err)
		}
	}

}

type IPCConn struct {
	ipc         ipc.IPC
	inputStream *ipc.ReadableStream
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
		log.Println("on request: ", request.ID)

		url, _ := url.ParseRequestURI(request.URL)

		//if (url.Host + url.Path) == "www.example.com/search" {
		if (url.Host + url.Path) == "127.0.0.1:8000/ipc/test" {
			//bodyReceiver := request.Body.(*ipc.BodyReceiver)
			//body := bodyReceiver.GetMetaBody().Data
			//log.Println("onRequest: ", request.URL, string(body), ic)

			body := `{"code": 0, "message": "hi"}`
			res := ipc.NewResponse(
				request.ID,
				200,
				ipc.NewHeaderWithExtra(map[string]string{
					"Content-Type": "application/json",
				}),
				ipc.NewBodySender([]byte(body), ic),
				ic,
			)

			if err := ic.PostMessage(context.TODO(), res); err != nil {
				log.Println("post message err: ", err)
			}
		}
	})

	inputStream := ipc.NewReadableStream()

	ipcConn := &IPCConn{
		ipc:         serverIPC,
		inputStream: inputStream,
		conn:        conn,
	}

	go func() {
		defer ipcConn.Close()
		// 读取inputStream数据并emit消息（接收消息并处理，然后把结果发送至内部流）
		if err := serverIPC.BindInputStream(inputStream); err != nil {
			panic(err)
		}
	}()

	go func() {
		// 读取内部流数据，然后response
		serverIPC.ReadOutputStream(func(data []byte) {
			if err := ipcConn.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				log.Println("res msg err: ", err)
				panic(err)
			}
		})
	}()

	return ipcConn
}

func (i *IPCConn) Close() {
	i.ipc.Close()
	i.inputStream.Controller.Close()
	i.conn.Close()
}
