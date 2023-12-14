package ipc

import (
	"errors"
	"io"
)

type BodyInter interface {
	Raw() any
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

func (b *Body) Raw() any {
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
	// TODO 完善
	return b.bodyHub.Stream
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
	bodyStream := b.Stream()
	if bodyStream == nil || len(p) == 0 {
		return 0, nil
	}

	reader := bodyStream.GetReader()
	for {
		r, err := reader.Read()
		if err != nil || r.Done {
			break
		}

		p = append(p, r.Value...)
	}

	return len(p), nil
}

func (b *Body) read(p []byte, src []byte) (int, error) {
	if b.offset >= len(src) {
		b.offset = 0
		return 0, io.EOF
	}

	n := copy(p, src[b.offset:])
	b.offset += n

	return n, nil
}

type BodyHub struct {
	Data   any             `json:"data"` // 类型是 string | []byte | ReadableStream
	Text   *string         `json:"text"`
	Stream *ReadableStream `json:"stream"`
	U8a    []byte          `json:"u8a"`
}

// NewBodyHub
// data类型只能是 string | []byte | ReadableStream
func NewBodyHub(data any) *BodyHub {
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
var metaIDReceiverIpcMap = make(map[string]IPC)

// TODO 待实现
// 每一个 metaBody 背后，都会有一个 IpcBodySender,
// 这里主要是存储 流，因为它有明确的 open/close 生命周期
var metaIDIpcBodySenderMap = make(map[string]*BodySender)

// TODO ts是用WeakMap实现的，这里暂时用map
// 任意的 RAW 背后都会有一个 IpcBodySender/IpcBodyReceiver
// 将它们缓存起来，那么使用这些 RAW 确保只拿到同一个 IpcBody，这对 RAW-Stream 很重要，流不可以被多次打开读取
var rawIpcBodyWMap = make(map[*ReadableStream]BodyInter)
