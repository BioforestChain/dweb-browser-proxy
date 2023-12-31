package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"path"
	"proxyServer/ipc"
	"time"
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
	res, err := ipcConn.ipc.Send(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	if res.StatusCode != 200 {
		log.Fatalln(res.StatusCode, string(res.Body.GetMetaBody().Data))
	}

	bodyStream := res.Body.Stream()
	if bodyStream == nil {
		log.Fatalln("bodyStream is nil")
	}

	data, err := readStreamWithTimeout(bodyStream, time.Duration(10))
	if err != nil {
		log.Fatalln("read body stream err: ", err)
	}

	filename := path.Join(downloadPath, "golang.png")
	if err := os.WriteFile(filename, data, 0666); err != nil {
		log.Fatalln(err)
	}
}

func readStreamWithTimeout(stream *ipc.ReadableStream, t time.Duration) ([]byte, error) {
	timer := time.NewTimer(t * time.Second)

	var timeout bool
	go func() {
		select {
		case <-timer.C:
			timeout = true
			stream.Controller.Close()
		}
	}()

	reader := stream.GetReader()
	data := make([]byte, 0)
	var readErr error
	for {
		r, err := reader.Read()
		if err != nil {
			readErr = err
			break
		}

		if r.Done {
			break
		}

		data = append(data, r.Value...)
	}

	if timeout {
		return nil, errors.New("read body stream timeout")
	}

	if readErr != nil {
		return nil, readErr
	}

	return data, nil
}

type IPCConn struct {
	ipc         ipc.IPC
	inputStream *ipc.ReadableStream
	conn        net.Conn
}

func NewIPCConn(conn net.Conn) *IPCConn {
	readableStreamIPC := ipc.NewReadableStreamIPC(ipc.CLIENT, ipc.SupportProtocol{
		Raw:         true,
		MessagePack: false,
		ProtoBuf:    false,
	})

	inputStream := ipc.NewReadableStream()

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
				log.Fatalln("Read: ", err)
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
