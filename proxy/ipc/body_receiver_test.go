package ipc

import (
	"bytes"
	"testing"
	"time"
)

func TestNewBodyReceiver(t *testing.T) {
	t.Run("metaBody is stream", func(t *testing.T) {
		metaBody1 := NewMetaBody(1, []byte("abc"), WithMetaBodyType(STREAM_WITH_BINARY), WithStreamID("s1"))
		ipc1 := NewReadableStreamIPC(CLIENT, SupportProtocol{})
		bodyReceiver := NewBodyReceiver(metaBody1, ipc1)
		if bodyReceiver.bodyHub.Stream == nil || bodyReceiver.metaBody.ReceiverUID != ipc1.GetUID() {
			t.Fatal("new BodyReceiver failed")
		}

		metaBody2 := NewMetaBody(1, []byte("abc"), WithMetaBodyType(STREAM_WITH_BINARY), WithStreamID("s1"))
		ipc2 := NewReadableStreamIPC(CLIENT, SupportProtocol{})
		_ = NewBodyReceiver(metaBody2, ipc2)
		if len(metaId_receiverIpc_Map) != 1 {
			t.Fatal("new BodyReceiver metaID_receiverIpc_Map cache failed")
		}

		ipc1.Close()
		// ugly block, because observers of signal are asynchronous executed
		time.Sleep(100 * time.Millisecond)
		if len(metaId_receiverIpc_Map) != 0 {
			t.Fatal("new BodyReceiver metaID_receiverIpc_Map delete cache failed")
		}

	})

	t.Run("metaBody is not stream", func(t *testing.T) {
		data := []byte("abc")
		metaBody := NewMetaBody(1, data, WithMetaBodyType(INLINE_BINARY), WithStreamID("s1"))
		ipc := NewReadableStreamIPC(CLIENT, SupportProtocol{})
		bodyReceiver := NewBodyReceiver(metaBody, ipc)

		if !bytes.Equal(bodyReceiver.bodyHub.U8a, data) {
			t.Fatal("new BodyReceiver failed")
		}
	})

}
