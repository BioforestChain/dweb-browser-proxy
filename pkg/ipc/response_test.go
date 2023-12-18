package ipc

import (
	"bytes"
	"net/http"
	"testing"
)

func TestFromResponseBinary(t *testing.T) {
	data := []byte("你好")
	ipc := NewReadableStreamIPC(SERVER, SupportProtocol{Raw: true})
	res := FromResponseBinary(1, http.StatusOK, NewHeader(), data, ipc)
	metaBody := res.Body.GetMetaBody()
	if metaBody.Type != INLINE_BINARY || !bytes.Equal(metaBody.Data, data) {
		t.Fatal("FromResponseBinary failed")
	}
}
