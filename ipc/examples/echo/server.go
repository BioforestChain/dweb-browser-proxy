////go:build ignore
//// +build ignore

package main

import (
	"context"
	"log"
	"net"
	"proxyServer/ipc"
	"time"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	ipcConn := NewIPCConn(conn)

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Println("Error reading:", err)
		return
	}

	_ = ipcConn.inputStream.Enqueue(buffer[:n])
}

type IPCConn struct {
	ipc         ipc.IPC
	inputStream *ipc.ReadableStream
	conn        net.Conn
}

func NewIPCConn(conn net.Conn) *IPCConn {
	serverIPC := ipc.NewReadableStreamIPC(ipc.SERVER, ipc.SupportProtocol{
		Raw:         true,
		MessagePack: false,
		ProtoBuf:    false,
	})

	// 监听并处理请求，echo请求数据
	serverIPC.OnRequest(func(req any, ic ipc.IPC) {
		request := req.(*ipc.Request)
		if request.URL == "https://www.example.com/search" && request.Method == ipc.POST {
			body := request.Body.GetMetaBody().Data
			log.Println("onRequest: ", time.Now(), request.ID, request.URL, string(body))

			// 处理request

			if err := ic.PostMessage(context.TODO(), request); err != nil {
				log.Println("post message err: ", err)
			}
		}
	})

	ipcConn := &IPCConn{
		ipc:         serverIPC,
		inputStream: ipc.NewReadableStream(),
		conn:        conn,
	}

	go func() {
		defer ipcConn.Close()
		// 读取inputStream数据并emit消息（接收消息并处理，然后把结果发送至输出流）
		if err := serverIPC.BindInputStream(ipcConn.inputStream); err != nil {
			panic(err)
		}
	}()

	go func() {
		// 读取输出流数据，然后response
		serverIPC.ReadOutputStream(func(data []byte) {
			if _, err := ipcConn.conn.Write(data); err != nil {
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
