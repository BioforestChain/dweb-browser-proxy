////go:build ignore
//// +build ignore

package main

import (
	"context"
	ipc2 "github.com/BioforestChain/dweb-browser-proxy/pkg/ipc"
	"log"
	"net"
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
	ipc         ipc2.IPC
	inputStream *ipc2.ReadableStream
	conn        net.Conn
}

func NewIPCConn(conn net.Conn) *IPCConn {
	serverIPC := ipc2.NewReadableStreamIPC(ipc2.SERVER, ipc2.SupportProtocol{
		Raw:         true,
		MessagePack: false,
		ProtoBuf:    false,
	})

	// 监听并处理请求，echo请求数据
	serverIPC.OnRequest(func(req any, ic ipc2.IPC) {
		request := req.(*ipc2.Request)
		if request.URL == "https://www.example.com/search" && request.Method == ipc2.POST {
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
		inputStream: ipc2.NewReadableStream(),
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
