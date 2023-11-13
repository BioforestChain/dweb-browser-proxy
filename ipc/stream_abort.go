package ipc

import "fmt"

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

func (s *StreamAbort) String() string {
	return fmt.Sprintf("StreamAbort - Type: %d, StreamID: %s", s.Type, s.StreamID)
}
