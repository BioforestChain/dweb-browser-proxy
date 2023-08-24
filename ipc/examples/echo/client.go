////go:build ignore
//// +build ignore

package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"proxyServer/ipc"
	"proxyServer/ipc/helper"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			do(&wg)
		}()
	}
	wg.Wait()
}

func do(wg *sync.WaitGroup) {
	defer wg.Done()
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	data := request()
	_, err = conn.Write(data)
	if err != nil {
		panic(err)
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil && err != io.EOF {
		panic(err)
	}

	log.Println("receiver data: ", string(buffer[4:n]))
}

func request() []byte {
	data := []byte("get my score")

	req := ipc.FromRequest(
		1,
		ipc.NewReadableStreamIPC(ipc.CLIENT, ipc.SupportProtocol{}),
		"https://www.example.com/search",
		ipc.RequestArgs{
			Method: "POST",
			Body:   data,
			Header: ipc.NewHeaderWithExtra(map[string]string{"Content-Type": "application/json"}),
		},
	)

	reqMarshal, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}

	return helper.FormatIPCData(reqMarshal)
}
