package ipc

import (
	"bytes"
	"encoding/json"
	"ipc/helper"
	"testing"
)

func TestReadableStreamIPC_postMessage(t *testing.T) {
	ipc := NewReadableStreamIPC(CLIENT, SupportProtocol{})
	url := "https://www.example.com"

	t.Run("request with empty body", func(t *testing.T) {
		req := ipc.Request(url, RequestArgs{Method: "GET"})
		err := ipc.postMessage(req)
		if err != nil {
			t.Fatal("readable stream ipc postMessage failed")
		}

		reqData := <-ipc.stream.GetReader().Read()

		var reqMessage ReqMessage
		err = json.Unmarshal(reqData[4:], &reqMessage)
		if err != nil || reqMessage.URL != url || reqMessage.MetaBody == nil {
			t.Fatal("readable stream ipc postMessage failed")
		}
	})

	t.Run("request with body", func(t *testing.T) {
		body := []byte("hi")
		req := ipc.Request(url, RequestArgs{Method: "GET", Body: body})
		err := ipc.postMessage(req)
		if err != nil {
			t.Fatal("readable stream ipc postMessage failed")
		}

		reqData := <-ipc.stream.GetReader().Read()

		//reqBytes, _ := json.Marshal(req)
		//msgLen := binary.LittleEndian.Uint32(reqData[0:4])
		//if msgLen != uint32(len(reqBytes)) {
		//	t.Fatal("readable stream ipc postMessage failed")
		//}

		var reqMessage ReqMessage
		err = json.Unmarshal(reqData[4:], &reqMessage)
		if err != nil || !bytes.Equal(reqMessage.MetaBody.Data, body) {
			t.Fatal("readable stream ipc postMessage failed")
		}
	})
}

func TestReadableStreamIPC_BindIncomeStream(t *testing.T) {
	t.Run("BindIncomeStream", func(t *testing.T) {
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
				t.Fatal("readable stream ipc bindincomestream failed")
			}

			body, ok := req.Body.(*BodyReceiver)
			if !ok {
				t.Fatal("readable stream ipc bindincomestream failed")
			}

			if !bytes.Equal(body.metaBody.Data, raw) {
				t.Fatal("readable stream ipc bindincomestream failed")
			}
		})

		proxyStream := NewReadableStream()
		go func() {
			if err := ipc.BindIncomeStream(proxyStream); err != nil {
				t.Error("readable stream ipc BindIncomeStream failed")
				return
			}
		}()

		dataEncoding := helper.U32To8Concat(uint32(len(data)), data)
		proxyStream.Controller.Enqueue(dataEncoding)

		<-ch
	})

	//t.Run("BindIncomeStream pong or close", func(t *testing.T) {
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
	//			t.Fatal("readable stream ipc bindincomestream failed")
	//		}
	//	})
	//
	//	proxyStream := NewReadableStream()
	//	go func() {
	//		if err := ipc.BindIncomeStream(proxyStream); err != nil {
	//			t.Error("readable stream ipc BindIncomeStream failed")
	//			return
	//		}
	//	}()
	//
	//	dataEncoding := helper.U32To8Concat(uint32(len(raw)), raw)
	//	proxyStream.Controller.Enqueue(dataEncoding)
	//
	//	<-ch
	//})
}

func Test_readIncomeStream(t *testing.T) {

	t.Run("readIncomeStream", func(t *testing.T) {
		raw := []byte("abcd")
		ch := make(chan struct{})
		var failed bool

		proxyStream := NewReadableStream()

		go func() {
			defer close(ch)
			for v := range readIncomeStream(proxyStream) {
				if !bytes.Equal(v, raw) {
					failed = true
					return
				}
			}
		}()

		dataEncoding := helper.U32To8Concat(uint32(len(raw)), raw)
		proxyStream.Controller.Enqueue(dataEncoding)
		// then close stream
		proxyStream.Controller.Close()

		<-ch

		if failed {
			t.Fatal("readIncomeStream failed")
		}
	})
}
