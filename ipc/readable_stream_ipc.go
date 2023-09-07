package ipc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"proxyServer/ipc/helper"
)

type ReadableStreamIPC struct {
	*BaseIPC
	role            ROLE
	supportProtocol SupportProtocol
	inputStream     *ReadableStream // 输入流
	outputStream    *ReadableStream // 输出流
}

func NewReadableStreamIPCWithDefaultInputStream(role ROLE, proto SupportProtocol) *ReadableStreamIPC {
	ipc := NewReadableStreamIPC(role, proto)

	ipc.inputStream = NewReadableStream()
	go func() {
		if err := ipc.bindInputStream(); err != nil {
			log.Println("ipc BindInputStream: ", err)
		}
	}()
	return ipc
}

func NewReadableStreamIPC(role ROLE, proto SupportProtocol) *ReadableStreamIPC {
	ipc := &ReadableStreamIPC{
		supportProtocol: proto,
		role:            role,
		outputStream:    NewReadableStream(),
	}

	ipc.BaseIPC = NewBaseIPC(
		WithPostMessage(ipc.postMessage),
		WithSupportProtocol(ipc.supportProtocol),
		WithDoClose(ipc.doClose),
		WithStreamReader(ipc.getOutputStreamReader),
	)

	return ipc
}

// BindInputStream reads messages from inputStream and emit data.
// A goroutine running BindInputStream is started for each inputStream. The
// application ensures that there is at most one reader on a stream by executing all
// reads from this goroutine.
// when calling inputStream.Controller.Close() or reading encounters an error, the read stops
func (rsi *ReadableStreamIPC) BindInputStream(inputStream *ReadableStream) (err error) {
	if rsi.inputStream != nil {
		return errors.New("income stream already bound")
	}
	rsi.inputStream = inputStream

	return rsi.bindInputStream()
}

func (rsi *ReadableStreamIPC) bindInputStream() (err error) {
	for data := range readInputStream(rsi.inputStream) {
		if len(data) == 4 || len(data) == 5 {
			switch {
			case helper.BytesEqual(data, "pong"):
				return nil
			case helper.BytesEqual(data, "close"):
				rsi.Close()
				return nil
			case helper.BytesEqual(data, "ping"):
				_ = rsi.outputStream.Enqueue(rsi.encode("ping"))
				return nil
			}
		}

		var msg interface{}

		if rsi.supportProtocol.MessagePack {
			panic("messagepack invalid")
			// TODO 其它编码
		} else {
			msg, err = objectToIpcMessage(data, rsi)
			if err != nil {
				return err
			}
		}

		helper.PutBuffer(data)

		rsi.msgSignal.Emit(msg, rsi)
	}

	// 输入流关闭后，输出流也要一起关闭
	rsi.Close()

	return nil
}

// msg类型：*Request | *Response | Event | StreamData | Stream*
func (rsi *ReadableStreamIPC) postMessage(ctx context.Context, msg interface{}) (err error) {
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

	select {
	case <-ctx.Done():
		err = ctx.Err()
	default:
		// 使用littleEndian存储msgLen
		chunk := helper.FormatIPCData(msgData)
		_ = rsi.outputStream.Enqueue(chunk)
	}

	return
}

// Enqueue write data to input stream
func (rsi *ReadableStreamIPC) Enqueue(data []byte) error {
	ipcData := helper.FormatIPCData(data)
	return rsi.inputStream.Enqueue(ipcData)
}

// ReadOutputStream 从输出流读取数据
func (rsi *ReadableStreamIPC) ReadOutputStream(cb func([]byte)) {
	reader := rsi.outputStream.GetReader()
	for {
		d, err := reader.Read()
		if err != nil || d.Done {
			return
		}

		cb(d.Value)
	}
}

// getOutputStreamReader 获取输出流channel
func (rsi *ReadableStreamIPC) getOutputStreamReader() *ReadableStreamDefaultReader {
	return rsi.outputStream.GetReader()
}

func (rsi *ReadableStreamIPC) doClose() {
	_ = rsi.outputStream.Enqueue(rsi.encode("close"))
	rsi.outputStream.Controller.Close()
}

func (rsi *ReadableStreamIPC) encode(msg string) []byte {
	// TODO 这里msg要根据数据传输协议encode
	return helper.FormatIPCData([]byte(msg))
}

// SupportProtocol 默认json？
type SupportProtocol struct {
	Raw, MessagePack, ProtoBuf bool
}

// readInputStream 会阻塞读取流数据，除非使用controller.Close()关闭流或发生错误
// 读取时，按 | header | body | header1 | body1 | ... | 顺序读取
// header由4字节组成，其uint32值，是body的大小
func readInputStream(stream *ReadableStream) <-chan []byte {
	var dataChan = make(chan []byte, 1)
	go func() {
		defer close(dataChan)
		b := newBinaryStreamRead(stream)

		for {
			// TODO 如果传输的数据未按 header|body格式传输，则bodySize一般会变得很大，导致一直阻塞读取
			// 需要校验格式
			header, err := b.read(4)
			if err != nil {
				break
			}

			bodySize := helper.GetBodySize(header)
			body, err := b.read(bodySize)
			if err != nil {
				break
			}

			helper.PutBuffer(header)

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
			if err != nil || v.Done {
				return
			}

			b.readChan <- v.Value
		}
	}()

	return b
}

// read 阻塞读取size byte
// TODO 读数据时需要加超时处理，如连接中断导致数据不全
func (b *binaryStreamRead) read(size int) ([]byte, error) {
	if b.cache.Len() >= size {
		v := b.cache.Next(size)
		c := helper.GetBuffer(len(v))
		copy(c, v)
		return c, nil
	}

	for v := range b.readChan {
		b.cache.Write(v)

		if b.cache.Len() >= size {
			v := b.cache.Next(size)
			c := helper.GetBuffer(len(v))
			copy(c, v)
			return c, nil
		}
	}

	return nil, io.EOF
}
