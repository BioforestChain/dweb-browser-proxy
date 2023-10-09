package ipc

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func Test_newStreamController(t *testing.T) {
	var i int
	var pauseCh = make(chan struct{})
	var abortCh = make(chan struct{})

	pulling := func(controller *streamController) {
		i++

		if i == 1 || i == 3 {
			pauseCh <- struct{}{}
		}
		time.Sleep(time.Millisecond * 20)
	}

	paused := func(controller *streamController) {

	}

	aborted := func(controller *streamController) {
		abortCh <- struct{}{}
	}

	sc := newStreamController(
		NewReadableStream(),
		WithStreamControllerPullingFunc(pulling),
		WithStreamControllerPauseFunc(paused),
		WithStreamControllerAbortFunc(aborted))

	// newStreamController会初始化goroutine，sc.Pulling会和goroutine通信
	// 避免sc.Pulling先执行，所以这里sleep等待goroutine初始化完
	time.Sleep(5 * time.Millisecond)
	sc.Pulling()

	<-pauseCh
	sc.Paused()

	if i != 1 {
		t.Fatal("newStreamController failed")
	}

	time.Sleep(time.Millisecond * 20)

	sc.Pulling()
	<-pauseCh
	if i != 3 {
		t.Fatal("newStreamController failed")
	}

	sc.Aborted()
	<-abortCh
	if i != 3 {
		t.Fatal("newStreamController failed")
	}
}

func Test_streamAsMeta(t *testing.T) {
	ipc := NewReadableStreamIPC(SERVER, SupportProtocol{})

	bodyStream := NewReadableStream()
	bs := NewBodySender(bodyStream, ipc)

	// 1. 模拟body stream数据
	go func() {
		_ = bodyStream.Enqueue([]byte("hi"))
	}()

	// 测试需要Sleep：NewBodySender存在接收信号的goroutine，这里不sleep的话
	// 会导致发送拉取信号操作先于goroutine初始化执行
	time.Sleep(10 * time.Millisecond)

	// 2. 发送拉取信号，读取bodyStream
	usedIpcInfoObj := bs.useByIpc(ipc)
	usedIpcInfoObj.emitStreamPull(NewStreamPulling(bodyStream.ID, nil))

	var ch = make(chan []byte)

	// 3. 从bodyStream里读取数据
	go func() {
		ipc.ReadOutputStream(func(data []byte) {
			ch <- data
		})
	}()

	data := <-ch

	var streamData StreamData
	_ = json.Unmarshal(data[4:], &streamData)

	if !bytes.Equal(streamData.Data, []byte("hi")) {
		t.Fatal("streamAsMeta falied")
	}

	usedIpcInfoObj.emitStreamAborted()

	time.Sleep(10 * time.Millisecond)

	if bs.usedIpcMap[usedIpcInfoObj.ipc] != nil {
		t.Fatal("streamAsMeta falied")
	}

	if metaIDIpcBodySenderMap[bs.metaBody.MetaID] != nil {
		t.Fatal("streamAsMeta falied")
	}
}

func Test_UsableByIpc(t *testing.T) {
	ipc := NewReadableStreamIPCWithDefaultInputStream(SERVER, SupportProtocol{})
	bodyStream := NewReadableStream()

	// 模拟body stream数据
	go func() {
		_ = bodyStream.Enqueue([]byte("hi"))
	}()

	bs := NewBodySender(bodyStream, ipc)
	// 1. 监听stream操作
	UsableByIpc(ipc, bs)

	time.Sleep(10 * time.Millisecond)

	// 2. 接收stream pulling请求
	streamPulling := NewStreamPulling(GetStreamID(bodyStream), nil)
	data, _ := json.Marshal(streamPulling)
	_ = ipc.EnqueueToInputStream(data)

	// 3. 读取返回的body stream数据
	var ch = make(chan []byte)
	go func() {
		ipc.ReadOutputStream(func(data []byte) {
			ch <- data
		})
	}()

	bodyStreamData := <-ch

	var streamData StreamData
	_ = json.Unmarshal(bodyStreamData[4:], &streamData)

	if !bytes.Equal(streamData.Data, []byte("hi")) {
		t.Fatal("streamAsMeta falied")
	}
}
