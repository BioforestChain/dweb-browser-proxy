package main

import (
	"context"
	"flag"
	ipc2 "github.com/BioforestChain/dweb-browser-proxy/pkg/ipc"
	"log"
	"net"
	"os"
	"path"
)

func main() {
	var downloadPath string
	flag.StringVar(&downloadPath, "o", "./", "path to save")
	flag.Parse()

	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ipcConn := NewIPCConn(conn)

	req := ipcConn.ipc.Request("https://www.example.com/golang/golang.png", ipc2.RequestArgs{
		Method: "GET",
		Body:   nil,
		Header: nil,
	})
	res, err := ipcConn.ipc.Send(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	if res.StatusCode != 200 {
		log.Fatalln(res.StatusCode, string(res.Body.GetMetaBody().Data))
	}

	filename := path.Join(downloadPath, "golang.png")
	if err := os.WriteFile(filename, res.Body.GetMetaBody().Data, 0666); err != nil {
		log.Fatalln(err)
	}

}

type IPCConn struct {
	ipc         ipc2.IPC
	inputStream *ipc2.ReadableStream
	conn        net.Conn
}

func NewIPCConn(conn net.Conn) *IPCConn {
	readableStreamIPC := ipc2.NewReadableStreamIPC(ipc2.CLIENT, ipc2.SupportProtocol{
		Raw:         true,
		MessagePack: false,
		ProtoBuf:    false,
	})

	inputStream := ipc2.NewReadableStream()

	ipcConn := &IPCConn{
		ipc:         readableStreamIPC,
		inputStream: inputStream,
		conn:        conn,
	}

	go func() {
		for {
			data := make([]byte, 512)
			n, err := ipcConn.conn.Read(data)
			if err != nil {
				log.Fatalln(err)
			}

			// 往输入流inputStream添加数据
			_ = ipcConn.inputStream.Enqueue(data[:n])
		}
	}()

	go func() {
		defer ipcConn.Close()
		// 读取输入流inputStream数据并emit消息（接收消息并处理，然后把结果发送至输出流）
		if err := readableStreamIPC.BindInputStream(inputStream); err != nil {
			panic(err)
		}
	}()

	go func() {
		// 读取输出流数据，然后response
		readableStreamIPC.ReadOutputStream(func(data []byte) {
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
