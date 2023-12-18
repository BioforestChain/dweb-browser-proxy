package main

import (
	"context"
	_ "embed"
	ipc2 "github.com/BioforestChain/dweb-browser-proxy/pkg/ipc"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
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

		ipcConn := NewIPCConn(conn)

		go handleConnection(ipcConn)
	}
}

func handleConnection(ipcConn *IPCConn) {
	conn := ipcConn.conn

	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			log.Println("Error reading:", err)
			return
		}

		_ = ipcConn.inputStream.Enqueue(buffer[:n])
	}
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
		if !(request.URL == "https://www.example.com/golang/golang.png" && request.Method == ipc2.GET) {
			return
		}

		var res *ipc2.Response
		path, _ := getAbsPath("ipc/examples/download-image/server/golang.png")
		f, err := os.Open(path)
		if err != nil {
			res = ipc2.FromResponseText(
				request.ID,
				http.StatusInternalServerError,
				ipc2.NewHeader(),
				http.StatusText(http.StatusInternalServerError),
				ic,
			)
		} else {
			bodyStream := ipc2.NewReadableStream()

			go func() {
				defer f.Close()
				for {
					data := make([]byte, 1024*10)
					n, err := f.Read(data)
					if err != nil {
						// close后，BodySender会发送stream end通知接收端
						bodyStream.Controller.Close()
						break
					}

					_ = bodyStream.Enqueue(data[:n])
				}
			}()

			res = ipc2.FromResponseStream(request.ID, http.StatusOK, ipc2.NewHeader(), bodyStream, ic)
		}

		if err := ic.PostMessage(context.TODO(), res); err != nil {
			log.Println("post message err: ", err)
		}
	})

	inputStream := ipc2.NewReadableStream()

	ipcConn := &IPCConn{
		ipc:         serverIPC,
		inputStream: inputStream,
		conn:        conn,
	}

	go func() {
		defer ipcConn.Close()
		// 读取inputStream数据并emit消息（接收消息并处理，然后把结果发送至输出流）
		if err := serverIPC.BindInputStream(inputStream); err != nil {
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

func getAbsPath(path string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Join(cwd, path), nil
}
