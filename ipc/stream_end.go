package ipc

import "fmt"

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

func (s *StreamEnd) String() string {
	return fmt.Sprintf("StreamEnd - Type: %d, StreamID: %s", s.Type, s.StreamID)
}
