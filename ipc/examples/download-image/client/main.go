package main

import (
	"flag"
	"log"
	"net"
	"os"
	"path"
	"proxyServer/ipc"
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

	req := ipcConn.ipc.Request("https://www.example.com/golang/golang.png", ipc.RequestArgs{
		Method: "GET",
		Body:   nil,
		Header: nil,
	})
	res, err := ipcConn.ipc.Send(req)
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
	ipc         ipc.IPC
	proxyStream *ipc.ReadableStream
	conn        net.Conn
}

func NewIPCConn(conn net.Conn) *IPCConn {
	readableStreamIPC := ipc.NewReadableStreamIPC(ipc.CLIENT, ipc.SupportProtocol{
		Raw:         true,
		MessagePack: false,
		ProtoBuf:    false,
	})

	proxyStream := ipc.NewReadableStream()

	ipcConn := &IPCConn{
		ipc:         readableStreamIPC,
		proxyStream: proxyStream,
		conn:        conn,
	}

	go func() {
		for {
			data := make([]byte, 512)
			n, err := ipcConn.conn.Read(data)
			if err != nil {
				log.Fatalln(err)
			}

			ipc.StreamDataEnqueue(ipcConn.proxyStream, data[:n])
		}
	}()

	go func() {
		defer ipcConn.Close()
		// 读取proxyStream数据并emit消息（接收消息并处理，然后把结果发送至输出流）
		if err := readableStreamIPC.BindIncomeStream(proxyStream); err != nil {
			panic(err)
		}
	}()

	go func() {
		// 读取输出流数据，然后response
		readableStreamIPC.ReadFromStream(func(data []byte) {
			if _, err := ipcConn.conn.Write(data); err != nil {
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
