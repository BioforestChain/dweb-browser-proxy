package ipc

import (
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"time"
)

var ErrUnMarshalObjectToIpcMessage = errors.New("unmarshal message failed when object to IpcMessage")

func objectToIpcMessage(data []byte, ipc IPC) (msg any, err error) {
	var m = struct {
		Type      MessageType  `json:"type"`
		Name      string       `json:"name"`
		Data      []byte       `json:"data"`
		Encoding  DataEncoding `json:"encoding"`
		StreamID  string       `json:"stream_id"`
		Bandwidth *int         `json:"bandwidth"`
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
		fmt.Printf("%s Input-> Request: %+v\n", time.Now(), msg)
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
		fmt.Printf("%s Input-> Response: %+v\n", time.Now(), msg)
	case STREAM_DATA:
		v := NewStreamData(m.StreamID, m.Data, m.Encoding)
		msg = v
		fmt.Printf("%s Input-> %+v\n", time.Now(), v)
	case STREAM_PULLING:
		msg = NewStreamPulling(m.StreamID, m.Bandwidth)
		fmt.Printf("%s Input-> %+v\n", time.Now(), msg)
	case STREAM_PAUSED:
		msg = NewStreamPaused(m.StreamID, m.Bandwidth)
		fmt.Printf("%s Input-> %+v\n", time.Now(), msg)
	case STREAM_ABORT:
		msg = NewStreamAbort(m.StreamID)
		fmt.Printf("%s Input-> %+v\n", time.Now(), msg)
	case STREAM_END:
		msg = NewStreamEnd(m.StreamID)
		fmt.Printf("%s Input-> %+v\n", time.Now(), msg)
	case EVENT:
		msg = NewEvent(m.Name, m.Data, m.Encoding)
	}

	return
}
