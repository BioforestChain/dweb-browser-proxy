package ipc

import "encoding/json"

type Response struct {
	ReqID      uint64
	StatusCode int
	Header     Header
	Body       interface{} // *Body | *BodySender | *BodyReceiver
	Ipc        IPC
	Type       MessageType
	resMessage *ResMessage
}

func NewResponse(reqID uint64, statusCode int, header Header, body interface{}, ipc IPC) *Response {
	if bodySender, ok := body.(*BodySender); ok {
		UsableByIpc(ipc, bodySender)
	}

	return &Response{Type: RESPONSE, ReqID: reqID, StatusCode: statusCode, Header: header, Body: body, Ipc: ipc}
}

func (res *Response) GetResMessage() *ResMessage {
	if res.resMessage == nil {
		var metaBody *MetaBody
		if res.Body != nil {
			switch v := res.Body.(type) {
			case *Body:
				metaBody = v.metaBody
			case *BodySender:
				metaBody = v.metaBody
			case *BodyReceiver:
				metaBody = v.metaBody
			}
		}
		res.resMessage = NewResMessage(res.ReqID, res.StatusCode, res.Header.toJSON(), metaBody)
	}
	return res.resMessage
}

func (res *Response) MarshalJSON() ([]byte, error) {
	reqMessage := res.GetResMessage()
	return json.Marshal(reqMessage)
}

type ResMessage struct {
	ReqID      uint64            `json:"req_id"`
	StatusCode int               `json:"statusCode"`
	Header     map[string]string `json:"header"`
	MetaBody   *MetaBody         `json:"metaBody"`
	Type       MessageType       `json:"type"`
}

func NewResMessage(reqID uint64, statusCode int, header map[string]string, body *MetaBody) *ResMessage {
	return &ResMessage{
		ReqID:      reqID,
		StatusCode: statusCode,
		Header:     header,
		MetaBody:   body,
		Type:       RESPONSE,
	}
}
