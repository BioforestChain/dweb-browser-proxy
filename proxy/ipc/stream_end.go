package ipc

type StreamEnd struct {
	Type     MessageType
	StreamID string
}

func NewStreamEnd(streamID string) *StreamEnd {
	return &StreamEnd{
		Type:     STREAM_END,
		StreamID: streamID,
	}
}
