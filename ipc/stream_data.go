package ipc

import (
	"fmt"
	"proxyServer/ipc/helper"
)

type StreamData struct {
	Type     MessageType
	StreamID string
	Data     []byte
	Encoding DataEncoding
}

func NewStreamData(streamID string, data []byte, encoding DataEncoding) *StreamData {
	return &StreamData{
		Type:     STREAM_DATA,
		StreamID: streamID,
		Data:     data,
		Encoding: encoding,
	}
}

func (s *StreamData) String() string {
	return fmt.Sprintf("StreamData - Type: %d, StreamID: %s, Data: %d, Encoding: %d", s.Type, s.StreamID, len(s.Data), s.Encoding)
}

func FromStreamDataBinary(streamID string, data []byte) *StreamData {
	return NewStreamData(streamID, data, BINARY)
}

type StreamMsg = StreamData

func IsStream(data any) (StreamMsg, bool) {
	switch v := data.(type) {
	case *StreamData:
		return StreamMsg{Type: v.Type, StreamID: v.StreamID, Data: v.Data, Encoding: v.Encoding}, true
	case *StreamPulling:
		return StreamMsg{Type: v.Type, StreamID: v.StreamID}, true
	case *StreamPaused:
		return StreamMsg{Type: v.Type, StreamID: v.StreamID}, true
	case *StreamEnd:
		return StreamMsg{Type: v.Type, StreamID: v.StreamID}, true
	case *StreamAbort:
		return StreamMsg{Type: v.Type, StreamID: v.StreamID}, true
	default:
		return StreamMsg{}, false
	}
}

func DataToBinary(data any, encoding DataEncoding) (r []byte) {
	switch encoding {
	case UTF8:
		r = data.([]byte)
	case BASE64:
		r, _ = helper.SimpleEncoder(data.(string), "base64")
	case BINARY:
		r = data.([]byte)
	}
	return
}
