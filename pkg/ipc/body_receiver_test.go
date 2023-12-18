package ipc

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestNewBodyReceiver(t *testing.T) {
	t.Run("metaBody is stream", func(t *testing.T) {
		metaBody1 := NewMetaBody(STREAM_WITH_BINARY, 1, []byte("abc"), WithStreamID("s1"))
		ipc1 := NewReadableStreamIPC(CLIENT, SupportProtocol{})
		bodyReceiver := NewBodyReceiver(metaBody1, ipc1)
		if bodyReceiver.bodyHub.Stream == nil || bodyReceiver.metaBody.ReceiverUID != ipc1.GetUID() {
			t.Fatal("new BodyReceiver failed")
		}

		metaBody2 := NewMetaBody(STREAM_WITH_BINARY, 1, []byte("abc"), WithStreamID("s1"))
		ipc2 := NewReadableStreamIPC(CLIENT, SupportProtocol{})
		_ = NewBodyReceiver(metaBody2, ipc2)
		if len(metaIDReceiverIpcMap) != 1 {
			t.Fatal("new BodyReceiver metaIDReceiverIpcMap cache failed")
		}

		ipc1.Close()
		// ugly block, because observers of signal are asynchronous executed
		time.Sleep(10 * time.Millisecond)
		if len(metaIDReceiverIpcMap) != 0 {
			t.Fatal("new BodyReceiver metaIDReceiverIpcMap delete cache failed")
		}

	})

	t.Run("metaBody is not stream", func(t *testing.T) {
		data := []byte("abc")
		metaBody := NewMetaBody(INLINE_BINARY, 1, data, WithStreamID("s1"))
		ipc := NewReadableStreamIPC(CLIENT, SupportProtocol{})
		bodyReceiver := NewBodyReceiver(metaBody, ipc)

		if !bytes.Equal(bodyReceiver.bodyHub.U8a, data) {
			t.Fatal("new BodyReceiver failed")
		}
	})
}

func Test_metaToStream(t *testing.T) {
	// mock server端body stream
	senderBodyStream := NewReadableStream()
	streamID := senderBodyStream.ID

	data := []byte("hi")
	mockBodyStreamData := FromStreamDataBinary(streamID, data)
	bodyStreamData, _ := json.Marshal(mockBodyStreamData)

	ipcClient := NewReadableStreamIPCWithDefaultInputStream(CLIENT, SupportProtocol{})
	metaBody := NewMetaBody(STREAM_ID, ipcClient.GetUID(), nil, WithStreamID(streamID))
	bodyReceiver := NewBodyReceiver(metaBody, ipcClient)

	result := make(chan *StreamResult)
	go func() {
		receiverBodyStream := bodyReceiver.bodyHub.Stream
		reader := receiverBodyStream.GetReader()

		// 1. read body stream时，会给server发stream pulling请求（下面ReadOutputStream读取的就是这个请求，
		// 并负责发送给server，这里模拟需要，ReadOutputStream内没有进行发送操作）
		// read读取的是下面ipcClient.EnqueueToInputStream的数据
		for {
			r, _ := reader.Read()
			result <- r
		}
	}()

	ch := make(chan struct{})
	// 2. client端发送pulling请求
	go func() {
		ipcClient.ReadOutputStream(func(data []byte) {
			var streamPulling StreamPulling
			_ = json.Unmarshal(data[4:], &streamPulling)
			if streamPulling.StreamID != "" {
				ch <- struct{}{}
			}
		})
	}()

	// 3. mock client端接口server发过来的body stream data（server端收到stream pulling请求后，会发送stream data）
	go func() {
		<-ch
		_ = ipcClient.EnqueueToInputStream(bodyStreamData)
	}()

	r := <-result
	if !bytes.Equal(r.Value, data) {
		t.Fatal("metaToStream failed")
	}
}
