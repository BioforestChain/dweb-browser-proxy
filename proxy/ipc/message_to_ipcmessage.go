package ipc

import (
	"encoding/json"
	"errors"
)

var ErrUnMarshalObjectToIpcMessage = errors.New("unmarshal message failed when object to IpcMessage")

func objectToIpcMessage(data []byte, ipc IPC) (msg interface{}, err error) {
	var m map[string]interface{}
	if err = json.Unmarshal(data, &m); err != nil {
		err = errors.Join(ErrUnMarshalObjectToIpcMessage, err)
		return
	}

	typ, ok := m["type"]
	if !ok {
		return nil, errors.New("data format is incorrect")
	}
	tp := int(typ.(float64))

	switch MessageType(tp) {
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
		msg = NewEvent(m["name"].(string), m["data"], m["encoding"].(DataEncoding))
	case STREAM_DATA:
		msg = NewStreamData(m["stream_id"].(string), m["data"], m["encoding"].(DataEncoding))
	case STREAM_PULLING:
		msg = NewStreamPulling(m["stream_id"].(string), m["bandwidth"].(*int))
	case STREAM_PAUSED:
		msg = NewStreamPaused(m["stream_id"].(string), m["bandwidth"].(*int))
	case STREAM_ABORT:
		msg = NewStreamAbort(m["stream_id"].(string))
	case STREAM_END:
		msg = NewStreamEnd(m["stream_id"].(string))
	}

	return
}
