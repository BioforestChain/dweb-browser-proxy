package ipc

type BodySender struct {
	Body
	isStream bool
	ipc      IPC
}

// NewBodySender data类型包括：string | []byte | *ReadableStream
func NewBodySender(data interface{}, ipc IPC) *BodySender {
	bs := &BodySender{ipc: ipc}
	bs.bodyHub = NewBodyHub(data)
	if bs.bodyHub.Stream != nil {
		bs.isStream = true
	}

	bs.metaBody = bodyAsMeta(data, ipc)

	// 作为 "生产者"，第一持有这个 IpcBodySender
	usableByIpc(ipc, bs)
	return bs
}

func bodyAsMeta(body interface{}, ipc IPC) *MetaBody {
	switch v := body.(type) {
	case string:
		return FromMetaBodyText(ipc.GetUID(), []byte(v))
	case []byte:
		return FromMetaBodyBinary(ipc, v)
	case *ReadableStream:
		return streamAsMeta(v, ipc)
	default:
		panic("bodyAsMeta body supports only string or ReadableStream or []byte")
	}
}

// TODO 完善
// 如果 rawData 是流模式，需要提供数据发送服务
// 这里不会一直无脑发，而是对方有需要的时候才发
func streamAsMeta(stream *ReadableStream, ipc IPC) *MetaBody {
	var streamFirstData []byte
	mb := NewMetaBody(ipc.GetUID(), streamFirstData, WithStreamID(getStreamID(stream)))
	mb.Type = STREAM_ID
	return mb
}

// usableByIpc 流数据监听及处理
func usableByIpc(ipc IPC, body *BodySender) {
	// TODO 待实现
}

func getStreamID(stream *ReadableStream) string {
	return stream.ID
}

func FromBodySenderText(data string, ipc IPC) *BodySender {
	return NewBodySender(data, ipc)
}

func FromBodySenderBinary(data []byte, ipc IPC) *BodySender {
	return NewBodySender(data, ipc)
}

func FromBodySenderStream(data *ReadableStream, ipc IPC) *BodySender {
	return NewBodySender(data, ipc)
}
