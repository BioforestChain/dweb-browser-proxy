package ipc

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestRequest_GetReqMessage(t *testing.T) {
	ipc := NewReadableStreamIPC(CLIENT, SupportProtocol{})
	bodySender := NewBodySender([]byte("abc"), ipc)
	req := NewRequest(1, "https://www.example.com", GET, Header{}, bodySender, ipc)

	reqMsg := req.GetReqMessage()
	if reqMsg.ReqID != 1 || reqMsg.MetaBody.Type != INLINE_BASE64 {
		t.Fatal("request GetReqMessage failed")
	}

	req1 := NewRequest(1, "https://www.example.com", GET, Header{}, nil, ipc)
	reqMsg1 := req1.GetReqMessage()
	if reqMsg1.MetaBody != nil {
		t.Fatal("request GetReqMessage failed")
	}
}

func TestRequest_MarshalJSON(t *testing.T) {
	ipc := NewReadableStreamIPC(CLIENT, SupportProtocol{})
	bodySender := NewBodySender([]byte("abc"), ipc)
	req := NewRequest(1, "https://www.example.com", GET, Header{}, bodySender, ipc)

	b, err := json.Marshal(req)
	if err != nil || !bytes.Contains(b, []byte("req_id")) {
		t.Fatal("request toJson failed")
	}
}
