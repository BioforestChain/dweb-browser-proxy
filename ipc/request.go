package ipc

import (
	"encoding/json"
	"fmt"
)

type Request struct {
	ID         uint64
	URL        string
	Method     METHOD
	Header     Header
	Body       BodyInter // *Body | *BodySender | *BodyReceiver
	Ipc        IPC
	Type       MessageType
	reqMessage *ReqMessage
}

// NewRequest body类型有*Body | *BodySender | *BodyReceiver
func NewRequest(id uint64, url string, method METHOD, header Header, body BodyInter, ipc IPC) *Request {
	req := &Request{
		ID:     id,
		URL:    url,
		Method: method,
		Header: header,
		Body:   body,
		Ipc:    ipc,
		Type:   REQUEST,
	}

	if bodySender, ok := body.(*BodySender); ok {
		UsableByIpc(ipc, bodySender)
	}

	return req
}

func (req *Request) GetReqMessage() *ReqMessage {
	if req.reqMessage != nil {
		return req.reqMessage
	}

	var metaBody *MetaBody
	if req.Body != nil {
		switch v := req.Body.(type) {
		case *Body:
			metaBody = v.metaBody
		case *BodySender:
			if v != nil {
				metaBody = v.metaBody
			}
		case *BodyReceiver:
			if v != nil {
				metaBody = v.metaBody
			}
		}
	}
	req.reqMessage = NewReqMessage(req.ID, req.Method, req.URL, req.Header.toJSON(), metaBody)

	return req.reqMessage
}

func (req *Request) MarshalJSON() ([]byte, error) {
	reqMessage := req.GetReqMessage()
	return json.Marshal(reqMessage)
}

func FromRequest(reqID uint64, ipc IPC, url string, init RequestArgs) *Request {
	var body *BodySender
	switch v := init.Body.(type) {
	case string:
		//body = FromBodySenderText(v, ipc)
		body = FromBodySenderBinary([]byte(v), ipc)
	case []byte:
		body = FromBodySenderBinary(v, ipc)
	case *ReadableStream:
		body = FromBodySenderStream(v, ipc)
	default:
		body = FromBodySenderText("", ipc)
	}

	return NewRequest(reqID, url, ToIPCMethod(init.Method), init.Header, body, ipc)
}

func FromRequestText(reqID uint64, url string, method METHOD, header Header, text string, ipc IPC) *Request {
	return NewRequest(reqID, url, method, header, FromBodySenderText(text, ipc), ipc)
}

func FromRequestBinary(reqID uint64, url string, method METHOD, header Header, data []byte, ipc IPC) *Request {
	header.init("Content-Type", "application/octet-stream")
	header.init("Content-Length", fmt.Sprintf("%d", len(data)))

	return NewRequest(reqID, url, method, header, FromBodySenderBinary(data, ipc), ipc)
}

func FromRequestStream(reqID uint64, url string, method METHOD, header Header, stream *ReadableStream, ipc IPC) *Request {
	header.init("Content-Type", "application/octet-stream")

	return NewRequest(reqID, url, method, header, FromBodySenderStream(stream, ipc), ipc)
}

type ReqMessage struct {
	ReqID    uint64            `json:"req_id"`
	Method   METHOD            `json:"method"`
	URL      string            `json:"url"`
	Header   map[string]string `json:"header"`
	MetaBody *MetaBody         `json:"metaBody"`
	Type     MessageType       `json:"type"`
}

func NewReqMessage(reqID uint64, method METHOD, url string, header map[string]string, metaBody *MetaBody) *ReqMessage {
	return &ReqMessage{
		ReqID:    reqID,
		Method:   method,
		URL:      url,
		Header:   header,
		MetaBody: metaBody,
		Type:     REQUEST,
	}
}
