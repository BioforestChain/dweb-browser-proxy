package ipc

import (
	"errors"
	jsoniter "github.com/json-iterator/go"
)

var ErrUnMarshalObjectToIpcMessage = errors.New("unmarshal message failed when object to IpcMessage")

func objectToIpcMessage(data []byte, ipc IPC) (msg interface{}, err error) {
	var m = struct {
		Type      MessageType
		Name      string
		Data      string
		Encoding  DataEncoding
		StreamID  string
		Bandwidth *int
	}{}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	if err = json.Unmarshal(data, &m); err != nil {
		err = errors.Join(ErrUnMarshalObjectToIpcMessage, err)
		return
	}

	switch m.Type {
	case REQUEST:
		var reqMsg ReqMessage
		if err = json.Unmarshal(data, &reqMsg); err != nil {
			return nil, err
		}
		msg = NewRequest(
			reqMsg.ReqID,
			reqMsg.URL,
			reqMsg.Method,
			NewHeaderWithExtra(reqMsg.Header),
			FromBodyReceiver(reqMsg.MetaBody, ipc),
			ipc,
		)
	case RESPONSE:
		var resMsg ResMessage
		if err = json.Unmarshal(data, &resMsg); err != nil {
			return nil, err
		}
		msg = NewResponse(
			resMsg.ReqID,
			resMsg.StatusCode,
			NewHeaderWithExtra(resMsg.Header),
			FromBodyReceiver(resMsg.MetaBody, ipc),
			ipc,
		)
	case EVENT:
		msg = NewEvent(m.Name, m.Data, m.Encoding)
	case STREAM_DATA:
		msg = NewStreamData(m.StreamID, m.Data, m.Encoding)
	case STREAM_PULLING:
		msg = NewStreamPulling(m.StreamID, m.Bandwidth)
	case STREAM_PAUSED:
		msg = NewStreamPaused(m.StreamID, m.Bandwidth)
	case STREAM_ABORT:
		msg = NewStreamAbort(m.StreamID)
	case STREAM_END:
		msg = NewStreamEnd(m.StreamID)
	}

	return
}
