package ipc

type StreamAbort struct {
	Type     MessageType
	StreamID string
}

func NewStreamAbort(streamID string) *StreamAbort {
	return &StreamAbort{
		Type:     STREAM_ABORT,
		StreamID: streamID,
	}
}
