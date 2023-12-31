package ipc

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestFromMetaBodyBinary(t *testing.T) {
	t.Run("sender is int without stream id", func(t *testing.T) {
		mb := FromMetaBodyBinary(uint64(1), []byte("hi"))
		if mb.Type != INLINE_BINARY {
			t.Fatal("FromMetaBodyBinary failed")
		}
	})

	t.Run("sender is int with stream id", func(t *testing.T) {
		mb := FromMetaBodyBinary(uint64(1), []byte("hi"), WithStreamID("abc"))
		if mb.Type != STREAM_WITH_BINARY {
			t.Fatal("FromMetaBodyBinary failed")
		}
	})

	t.Run("sender is ipc that support binary", func(t *testing.T) {
		ipc := NewBaseIPC(WithSupportProtocol(SupportProtocol{MessagePack: true}))

		mb := FromMetaBodyBinary(ipc, []byte("hi"))
		if mb.Type != INLINE_BINARY {
			t.Fatal("FromMetaBodyBinary failed")
		}
	})

	t.Run("sender is ipc that unsupported binary", func(t *testing.T) {
		ipc := NewBaseIPC(WithSupportProtocol(SupportProtocol{}))
		data := []byte("hi")

		mb := FromMetaBodyBinary(ipc, data)
		if mb.Type != INLINE_BASE64 {
			t.Fatal("FromMetaBodyBinary failed")
		}

		if !bytes.Equal(data, mb.Data) {
			t.Fatal("FromMetaBodyBinary failed")
		}
	})
}

func TestMetaBody_typeEncoding(t *testing.T) {
	assertTypeEncoding(t, INLINE, UTF8)
	assertTypeEncoding(t, STREAM_WITH_BASE64, BASE64)
	assertTypeEncoding(t, INLINE_BINARY, BINARY)
}

func assertTypeEncoding(t *testing.T, v MetaBodyType, wanted DataEncoding) {
	t.Helper()
	mb := NewMetaBody(1, []byte("abc"), WithMetaBodyType(v))
	got := mb.typeEncoding()
	if got != wanted {
		t.Fatal("MetaBody typeEncoding failed")
	}
}

func TestMetaBody_MarshalJSON(t *testing.T) {
	t.Run("", func(t *testing.T) {
		m := NewMetaBody(1, []byte("hi"), WithMetaBodyType(INLINE_TEXT))
		mb, err := json.Marshal(m)
		if err != nil {
			t.Fatal("metaBody marshal failed")
			return
		}

		if !bytes.Contains(mb, []byte(`"type":3`)) {
			t.Fatal("metaBody marshal failed")
		}
	})

	t.Run("", func(t *testing.T) {
		m := NewMetaBody(1, []byte("hi"), WithMetaBodyType(INLINE_BINARY))
		mb, err := json.Marshal(m)
		if err != nil {
			t.Fatal("metaBody marshal failed")
			return
		}

		if !bytes.Contains(mb, []byte(`"type":5`)) {
			t.Fatal("metaBody marshal failed")
		}
	})

}
