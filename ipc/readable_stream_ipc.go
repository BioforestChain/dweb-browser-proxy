package ipc

import (
	"bytes"
	"encoding/json"
	"errors"
	"proxyServer/ipc/helper"
)

type ReadableStreamIPC struct {
	*BaseIPC
	role            ROLE
	supportProtocol SupportProtocol
	stream          *ReadableStream // 输出流 TODO 需要考虑并发使用问题
	proxyStream     *ReadableStream // 输入流
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
		WithStreamReader(ipc.getStreamReader),
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
				_ = rsi.stream.Enqueue(rsi.encode("ping"))
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

	// 输入流关闭后，输出流也要一起关闭
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

	// 使用littleEndian存储msgLen
	chunk := helper.FormatIPCData(msgData)
	_ = rsi.stream.Enqueue(chunk)
	return
}

// ReadFromStream 从输出流读取数据
func (rsi *ReadableStreamIPC) ReadFromStream(cb func([]byte)) {
	reader := rsi.stream.GetReader()
	for {
		d, err := reader.Read()
		if err != nil {
			return
		}

		cb(d.Value)
	}
}

// getStreamReader 获取输出流channel
func (rsi *ReadableStreamIPC) getStreamReader() *ReadableStreamDefaultReader {
	return rsi.stream.GetReader()
}

func (rsi *ReadableStreamIPC) doClose() {
	_ = rsi.stream.Enqueue(rsi.encode("close"))
	rsi.stream.Controller.Close()
}

func (rsi *ReadableStreamIPC) encode(msg string) []byte {
	// TODO 这里msg要根据数据传输协议encode
	return helper.FormatIPCData([]byte(msg))
}

// SupportProtocol 默认json？
type SupportProtocol struct {
	Raw, MessagePack, ProtoBuf bool
}

// readIncomeStream 会一直读取流数据，除非使用controller.Close()关闭流或发生错误
// 读取时，按 | header | body | header1 | body1 | ... | 顺序读取
// header由4字节组成，其uint32值，是body的大小
func readIncomeStream(stream *ReadableStream) <-chan []byte {
	var dataChan = make(chan []byte)
	go func() {
		defer close(dataChan)
		b := newBinaryStreamRead(stream)
		for {
			// TODO 如果传输的数据未按 header|body格式传输，则读取的数据会出问题
			// 需要校验格式
			header := b.read(4)
			if header == nil {
				break
			}

			bodySize := helper.GetBodySize(header)
			body := b.read(bodySize)
			if body == nil {
				break
			}

			dataChan <- body
		}
	}()

	return dataChan
}

type binaryStreamRead struct {
	stream   *ReadableStream
	readChan chan []byte
	cache    *bytes.Buffer
}

func newBinaryStreamRead(stream *ReadableStream) *binaryStreamRead {
	b := &binaryStreamRead{
		stream:   stream,
		readChan: make(chan []byte, 1),
		cache:    new(bytes.Buffer),
	}

	go func() {
		defer func() {
			close(b.readChan)
			b.cache.Reset()
		}()

		reader := b.stream.GetReader()
		for {
			v, err := reader.Read()
			if err != nil {
				return
			}
			b.readChan <- v.Value
		}
	}()

	return b
}

// read 阻塞读取size byte
// TODO 读数据时需要加超时处理，如连接中断导致数据不全
func (b *binaryStreamRead) read(size int) []byte {
	if b.cache.Len() >= size {
		return b.cache.Next(size)
	}

	for v := range b.readChan {
		b.cache.Write(v)

		if b.cache.Len() >= size {
			return b.cache.Next(size)
		}
	}

	return nil
}
