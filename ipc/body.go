package ipc

type BodyInter interface {
	Raw() interface{}
	U8a() []byte
	Stream() *ReadableStream
	Text() string
	GetMetaBody() *MetaBody
}

type Body struct {
	metaBody *MetaBody
	bodyHub  *BodyHub
}

func (b *Body) Raw() interface{} {
	return b.bodyHub.Data
}

func (b *Body) U8a() []byte {
	bodyU8a := b.bodyHub.U8a
	if len(bodyU8a) == 0 {
		if b.bodyHub.Text != "" {
			bodyU8a = []byte(b.bodyHub.Text)
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
	if bodyText == "" {
		// TODO read from U8a
	}
	return bodyText
}

func (b *Body) GetMetaBody() *MetaBody {
	return b.metaBody
}

type BodyHub struct {
	Data   interface{} // 类型是 string | []byte | ReadableStream
	Text   string
	Stream *ReadableStream
	U8a    []byte
}

// NewBodyHub
// data类型只能是 string | []byte | ReadableStream
func NewBodyHub(data interface{}) *BodyHub {
	bh := &BodyHub{Data: data}
	switch v := data.(type) {
	case string:
		bh.Text = v
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
