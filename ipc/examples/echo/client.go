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
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	data := request(1)
	_, err = conn.Write(data)
	if err != nil {
		panic(err)
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil && err != io.EOF {
		panic(err)
	}

	log.Println("response: ", string(buffer[:n]))
}

func request(reqID int) []byte {
	data := []byte("get my score")

	req := ipc.FromRequest(
		uint64(reqID),
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
