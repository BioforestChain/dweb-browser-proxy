package ipc

import (
	"errors"
	"io"
)

type BodyInter interface {
	Raw() interface{}
	Text() string
	U8a() []byte
	Stream() *ReadableStream
	GetMetaBody() *MetaBody
	Read(p []byte) (int, error)
}

type Body struct {
	metaBody *MetaBody
	bodyHub  *BodyHub
	offset   int // used by Read body
}

func (b *Body) Raw() interface{} {
	return b.bodyHub.Data
}

func (b *Body) U8a() []byte {
	bodyU8a := b.bodyHub.U8a
	if bodyU8a == nil {
		if b.bodyHub.Text != nil {
			bodyU8a = []byte(*b.bodyHub.Text)
		} else if b.bodyHub.Stream != nil {
			// TODO read from stream
		} else {
			panic("invalid body type")
		}
	}

	return bodyU8a
}

func (b *Body) Stream() *ReadableStream {
	// TODO
	return nil
}

func (b *Body) Text() string {
	bodyText := b.bodyHub.Text
	if bodyText == nil {
		// TODO read from U8a
	}
	return *bodyText
}

func (b *Body) GetMetaBody() *MetaBody {
	return b.metaBody
}

func (b *Body) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	if b.bodyHub.Text != nil {
		return b.readText(p)
	}

	if b.bodyHub.U8a != nil {
		return b.readU8a(p)
	}

	if b.bodyHub.Stream != nil {
		return b.readReadableStream(p)
	}

	return 0, errors.New("unsupported data type")
}

func (b *Body) readText(p []byte) (int, error) {
	return b.read(p, []byte(*b.bodyHub.Text))
}

func (b *Body) readU8a(p []byte) (int, error) {
	return b.read(p, b.bodyHub.U8a)
}

func (b *Body) readReadableStream(p []byte) (n int, err error) {
	return 0, nil
}

func (b *Body) read(p []byte, data []byte) (int, error) {
	if b.offset >= len(data) {
		b.offset = 0
		return 0, io.EOF
	}

	n := copy(p, data[b.offset:])
	b.offset += n

	return n, nil
}

type BodyHub struct {
	Data   interface{} // 类型是 string | []byte | ReadableStream
	Text   *string
	Stream *ReadableStream
	U8a    []byte
}

// NewBodyHub
// data类型只能是 string | []byte | ReadableStream
func NewBodyHub(data interface{}) *BodyHub {
	bh := &BodyHub{Data: data}
	switch v := data.(type) {
	case string:
		bh.Text = &v
	case []byte:
		bh.U8a = v
	case *ReadableStream:
		bh.Stream = v
	default:
		panic("bodyhub supports only string, ReadableStream and []byte types")
	}

	return bh
}

// 每一个 metaBody 背后，都会有第一个 接收者IPC，这直接定义了它的应该由谁来接收这个数据，
// 其它的 IPC 即便拿到了这个 metaBody 也是没有意义的，除非它是 INLINE
var metaId_receiverIpc_Map = make(map[string]IPC)

// 每一个 metaBody 背后，都会有一个 IpcBodySender,
// 这里主要是存储 流，因为它有明确的 open/close 生命周期
var metaId_ipcBodySender_Map = make(map[string]*BodySender)
