package ipc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"proxyServer/ipc/helper"
	"testing"
	"time"
)

func TestReadableStreamIPC_NewReadableStreamIPCWithDefaultInputStream(t *testing.T) {
	readableStreamIpc := NewReadableStreamIPCWithDefaultInputStream(CLIENT, SupportProtocol{Raw: true})
	go func() {
		readableStreamIpc.ReadOutputStream(func(data []byte) {
			fmt.Println("data: ", string(data))
		})
	}()

	req := request()

	_ = readableStreamIpc.EnqueueToInputStream(req)
}

func TestReadableStreamIPC_postMessage(t *testing.T) {
	t.Run("request with empty body", func(t *testing.T) {
		ipc := NewReadableStreamIPC(CLIENT, SupportProtocol{})
		url := "https://www.example.com"

		req := ipc.Request(url, RequestArgs{Method: "GET"})
		err := ipc.postMessage(context.TODO(), req)
		if err != nil {
			t.Fatal("readable stream ipc postMessage failed")
		}

		reqData, _ := ipc.outputStream.GetReader().Read()

		var reqMessage ReqMessage
		err = json.Unmarshal(reqData.Value[4:], &reqMessage)
		if err != nil || reqMessage.URL != url || reqMessage.MetaBody == nil {
			t.Fatal("readable stream ipc postMessage failed")
		}
	})

	t.Run("request with body", func(t *testing.T) {
		ipc := NewReadableStreamIPC(CLIENT, SupportProtocol{})
		url := "https://www.example.com"

		body := []byte("hi")
		req := ipc.Request(url, RequestArgs{Method: "GET", Body: body})
		err := ipc.postMessage(context.TODO(), req)
		if err != nil {
			t.Fatal("readable stream ipc postMessage failed")
		}

		reqData, _ := ipc.outputStream.GetReader().Read()

		//reqBytes, _ := json.Marshal(req)
		//msgLen := binary.LittleEndian.Uint32(reqData[0:4])
		//if msgLen != uint32(len(reqBytes)) {
		//	t.Fatal("readable stream ipc postMessage failed")
		//}

		var reqMessage ReqMessage
		err = json.Unmarshal(reqData.Value[4:], &reqMessage)
		if err != nil || !bytes.Equal(reqMessage.MetaBody.Data, body) {
			t.Fatal("readable stream ipc postMessage failed")
		}
	})
}

func TestReadableStreamIPC_BindInputStream(t *testing.T) {
	t.Run("BindInputStream", func(t *testing.T) {
		ch := make(chan struct{})
		raw := []byte("abcd")
		ipc := NewReadableStreamIPC(CLIENT, SupportProtocol{})
		req := FromRequest(1, ipc, "https://www.example.com", RequestArgs{
			Method: "get",
			Body:   raw,
			Header: NewHeader(),
		})

		data, _ := json.Marshal(req)

		ipc.OnMessage(func(msg interface{}, ipc IPC) {
			defer close(ch)

			req, ok := msg.(*Request)
			if !ok {
				t.Fatal("readable stream ipc BindInputStream failed")
			}

			body, ok := req.Body.(*BodyReceiver)
			if !ok {
				t.Fatal("readable stream ipc BindInputStream failed")
			}

			if !bytes.Equal(body.metaBody.Data, raw) {
				t.Fatal("readable stream ipc BindInputStream failed")
			}
		})

		inputStream := NewReadableStream()
		go func() {
			if err := ipc.BindInputStream(inputStream); err != nil {
				t.Error("readable stream ipc BindInputStream failed")
				return
			}
		}()

		dataEncoding := helper.FormatIPCData(data)
		_ = inputStream.Enqueue(dataEncoding)

		<-ch
	})

	//t.Run("BindInputStream pong or close", func(t *testing.T) {
	//	raw := []byte("pong")
	//	ch := make(chan struct{})
	//
	//	ipc := NewReadableStreamIPC(CLIENT, &SupportProtocol{})
	//	ipc.OnMessage(func(req interface{}, ipc IPC) {
	//		defer close(ch)
	//
	//		fmt.Println("req: ", req, ipc)
	//
	//		if !bytes.Equal(req.([]byte), raw) {
	//			t.Fatal("readable stream ipc BindInputStream failed")
	//		}
	//	})
	//
	//	inputStream := NewReadableStream()
	//	go func() {
	//		if err := ipc.BindInputStream(inputStream); err != nil {
	//			t.Error("readable stream ipc BindInputStream failed")
	//			return
	//		}
	//	}()
	//
	//	dataEncoding := helper.U32To8Concat(uint32(len(raw)), raw)
	//	inputStream.Controller.Enqueue(dataEncoding)
	//
	//	<-ch
	//})
}

func Test_readInputStream(t *testing.T) {

	t.Run("readInputStream", func(t *testing.T) {
		raw := []byte("abcd")
		ch := make(chan struct{})
		var failed = true

		inputStream := NewReadableStream()

		go func() {
			defer close(ch)
			for v := range readInputStream(inputStream) {
				if bytes.Equal(v, raw) {
					failed = false
					return
				}
			}
		}()

		dataEncoding := helper.FormatIPCData(raw)
		_ = inputStream.Enqueue(dataEncoding)
		// then close stream
		time.Sleep(time.Millisecond * 10)
		inputStream.Controller.Close()

		<-ch

		if failed {
			t.Fatal("readInputStream failed")
		}
	})

	t.Run("readInputStream with consecutive enqueue", func(t *testing.T) {
		raw := []byte("中文abcdefghijklmnopqrstuvwsyzABCDEFGHIJKLMNOPQRSTUVWSYZ0123456789")
		dataEncoding := helper.FormatIPCData(raw)

		inputStream := NewReadableStream()

		ch := make(chan struct{})
		var failed = true

		go func() {
			defer close(ch)
			for v := range readInputStream(inputStream) {
				if bytes.Equal(v, raw) {
					failed = false
					return
				}
			}
		}()

		for _, b := range dataEncoding {
			_ = inputStream.Enqueue([]byte{b})
		}

		// then close stream
		time.Sleep(time.Millisecond * 10)
		inputStream.Controller.Close()

		<-ch

		if failed {
			t.Fatal("readInputStream failed")
		}
	})
}

func request() []byte {
	data := []byte("hi")

	req := FromRequest(
		1,
		NewReadableStreamIPC(CLIENT, SupportProtocol{}),
		"https://www.example.com/search",
		RequestArgs{
			Method: "POST",
			Body:   data,
			Header: NewHeaderWithExtra(map[string]string{"Content-Type": "application/json"}),
		},
	)

	reqMarshal, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}

	return reqMarshal
}
