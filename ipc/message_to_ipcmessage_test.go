package ipc

import (
	"encoding/json"
	"testing"
)

func Test_objectToIpcMessage(t *testing.T) {
	var url = "https://www.example.com"
	var reqID uint64 = 1
	var method = GET
	var senderUID uint64 = 2
	var data = []byte("abc")
	var ipc = NewReadableStreamIPC(CLIENT, SupportProtocol{})

	t.Run("object to request", func(t *testing.T) {
		metaBody := NewMetaBody(senderUID, data, WithMetaBodyType(INLINE_BINARY))
		reqMsg := NewReqMessage(reqID, method, url, map[string]string{}, metaBody)
		reqData, _ := json.Marshal(reqMsg)

		gotMsg, err := objectToIpcMessage(reqData, ipc)
		if err != nil {
			t.Fatalf("objectToIpcMessage() error = %v", err)
		}

		req, ok := gotMsg.(*Request)
		if !ok {
			t.Fatal("objectToIpcMessage to request failed")
		}

		bodyReceiver := req.Body.(*BodyReceiver)
		if bodyReceiver.metaBody.Type != INLINE_BASE64 {
			t.Fatal("objectToIpcMessage to request failed")
		}
	})

	t.Run("object to response", func(t *testing.T) {
		metaBody := NewMetaBody(senderUID, data, WithMetaBodyType(INLINE_TEXT))
		resMsg := NewResMessage(reqID, 200, map[string]string{}, metaBody)
		resData, _ := json.Marshal(resMsg)

		gotMsg, err := objectToIpcMessage(resData, ipc)
		if err != nil {
			t.Fatalf("objectToIpcMessage() error = %v", err)
		}

		res, ok := gotMsg.(*Response)
		if !ok {
			t.Fatal("objectToIpcMessage to response failed")
		}

		bodyReceiver := res.Body.(*BodyReceiver)
		if bodyReceiver.metaBody.Type != INLINE_TEXT {
			t.Fatal("objectToIpcMessage to response failed")
		}
	})
}
