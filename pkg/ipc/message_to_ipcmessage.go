package ipc

import (
	"errors"
	"fmt"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/ipc/helper"
	jsoniter "github.com/json-iterator/go"
	"time"
)

var ErrUnMarshalObjectToIpcMessage = errors.New("unmarshal message failed when object to IpcMessage")

func objectToIpcMessage(data []byte, ipc IPC) (msg any, err error) {
	var m = struct {
		Type      MessageType  `json:"type"`
		Name      string       `json:"name"`
		Data      any          `json:"data"`
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
		v := NewStreamData(m.StreamID, dataToBytes(m.Data, m.Encoding), m.Encoding)
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
		msg = NewEvent(m.Name, dataToBytes(m.Data, m.Encoding), m.Encoding)
	}

	return
}

// Proxy output:  后端发给前端格式{data: string, encoding: utf8}
// Proxy input: 前端数据格式有2种，json和cbor
//  1. 前端格式{data: text, encoding: UTF8} -> 对应后端格式{data: []byte, encoding: UTF8}
//  2. 前端格式{data: base64, encoding: BASE64} -> 对应后端格式{data: []byte, encoding: BASE64}
//     前端要发送binary数据，要使用CBOR格式，不然用json发送二进制数据会导致数据无法decode
//  3. 前端格式{data: binary, encoding: BINARY} -> 对应后端格式{data: []byte, encoding: BINARY} ?
func dataToBytes(data any, encoding DataEncoding) (r []byte) {
	switch encoding {
	case UTF8:
		r = []byte(data.(string))
	case BASE64:
		r, _ = helper.SimpleEncoder(data.(string), "base64")
	case BINARY:
		// TODO 服务端待实现CBOR的codec
		r, _ = helper.SimpleEncoder(data.(string), "base64")
	}
	return
}
