package ipc

type StreamData struct {
	Type     MessageType
	StreamID string
	Data     interface{} // string | []byte
	Encoding DataEncoding
}

func NewStreamData(streamID string, data interface{}, encoding DataEncoding) *StreamData {
	return &StreamData{
		Type:     STREAM_DATA,
		StreamID: streamID,
		Data:     data,
		Encoding: encoding,
	}
}
