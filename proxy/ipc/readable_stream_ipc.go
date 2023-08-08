package ipc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"ipc/helper"
	"log"
)

type ReadableStreamIPC struct {
	*BaseIPC
	role            ROLE
	supportProtocol SupportProtocol
	stream          *ReadableStream // 内部流 TODO 需要考虑并发使用问题
	proxyStream     *ReadableStream // 代理流
}

func NewReadableStreamIPC(role ROLE, proto SupportProtocol) *ReadableStreamIPC {
	ipc := &ReadableStreamIPC{
		supportProtocol: proto,
		role:            role,
		stream:          NewReadableStream(),
	}

	ipc.BaseIPC = NewBaseIPC(
		WithPostMessage(ipc.postMessage),
		WithSupportProtocol(ipc.supportProtocol),
		WithDoClose(ipc.doClose),
		WithStreamRead(ipc.getStreamRead),
	)

	return ipc
}

// BindIncomeStream reads messages from proxyStream and emit data.
// A goroutine running BindIncomeStream is started for each proxyStream. The
// application ensures that there is at most one reader on a stream by executing all
// reads from this goroutine.
// when calling proxyStream.Controller.Close() or reading encounters an error, the read stops
func (rsi *ReadableStreamIPC) BindIncomeStream(proxyStream *ReadableStream) (err error) {
	if rsi.proxyStream != nil {
		return errors.New("income stream already bound")
	}
	rsi.proxyStream = proxyStream

	for data := range readIncomeStream(rsi.proxyStream) {
		if len(data) == 4 || len(data) == 5 {
			switch {
			case helper.BytesEqual(data, "pong"):
				return nil
			case helper.BytesEqual(data, "close"):
				rsi.Close()
				return nil
			case helper.BytesEqual(data, "ping"):
				rsi.stream.Controller.Enqueue(rsi.encode("ping"))
				return nil
			}
		}

		var msg interface{}

		if rsi.supportProtocol.MessagePack {
			// TODO 其它编码
		} else {
			msg, err = objectToIpcMessage(data, rsi)
			if err != nil {
				return err
			}
		}

		rsi.msgSignal.Emit(msg, rsi)
	}

	// 代理流关闭后，内部流也要一起关闭
	rsi.Close()

	return nil
}

// msg类型：*Request | *Response | Event | StreamData | Stream*
func (rsi *ReadableStreamIPC) postMessage(msg interface{}) (err error) {
	var msgRaw interface{}
	switch v := msg.(type) {
	case *Request:
		msgRaw = v.GetReqMessage() // ReqMessage
	case *Response:
		msgRaw = v.GetResMessage() // ResMessage
	default:
		msgRaw = msg
	}

	var msgData []byte
	if rsi.supportProtocol.MessagePack {
		// TODO encode msgRaw use message pack
	} else {
		msgData, err = json.Marshal(msgRaw)
		if err != nil {
			return
		}
	}

	msgLen := uint32(len(msgData))
	// 使用littleEndian存储msgLen
	chunk := helper.U32To8Concat(msgLen, msgData)
	rsi.stream.Controller.Enqueue(chunk)
	return
}

// ReadFromStream 从内部流读取数据
func (rsi *ReadableStreamIPC) ReadFromStream(cb func([]byte)) {
	for data := range rsi.stream.GetReader().Read() {
		cb(data)
	}
}

// getStreamRead 获取内部流channel
func (rsi *ReadableStreamIPC) getStreamRead() <-chan []byte {
	return rsi.stream.GetReader().Read()
}

func (rsi *ReadableStreamIPC) doClose() {
	rsi.stream.Controller.Enqueue(rsi.encode("close"))
	rsi.stream.Controller.Close()
}

func (rsi *ReadableStreamIPC) encode(msg string) []byte {
	// TODO 这里msg要根据数据传输协议encode
	return helper.U32To8Concat(uint32(len(msg)), []byte(msg))
}

// SupportProtocol 默认json？
type SupportProtocol struct {
	Raw, MessagePack, ProtoBuf bool
}

// readIncomeStream 会一直读取流数据，除非使用controller.Close()关闭流或发生错误
// 读取时，按 | header | body | header1 | body1 | ... | 顺序读取
// header由4字节组成，其uint32值，是body的大小
func readIncomeStream(stream *ReadableStream) <-chan []byte {
	var closeChan bool
	var cache = new(bytes.Buffer)
	var reqChan = make(chan []byte)
	go func() {
		defer close(reqChan)
		defer cache.Reset()

		for {
			// stream closed
			if closeChan {
				break
			}

			select {
			case <-stream.CloseChan:
				fmt.Println("close stream chan")
				closeChan = true
				break
			case v := <-stream.GetReader().Read():
				cache.Write(v)
			}

			// 由于controller.Enqueue是把header和body一起入队，所以不存在一次流不完整情况？
			//if cache.Len() < 4 {
			//	continue
			//}

			// TODO 优化字节数组创建,使用sync.pool?
			header := make([]byte, 4)
			if _, err := cache.Read(header); err != nil {
				if err == io.EOF {
					continue
				}
				log.Println("read header error: ", err)
				return
			}

			// TODO 如果传输的数据未按 header|body格式传输，则读取的数据会出问题
			// 需要校验格式
			size := helper.U8aToU32(header)
			//if cache.Len() < int(size) {
			//	continue
			//}

			// TODO 优化：如果size太大，可以拆分成小段读取
			body := make([]byte, size)
			if _, err := cache.Read(body); err != nil {
				if err == io.EOF {
					continue
				}
				log.Println("read body error: ", err)
				return
			}

			reqChan <- body
		}
	}()

	return reqChan
}
