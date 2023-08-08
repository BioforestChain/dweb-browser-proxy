package ipc

import (
	"crypto/rand"
	"encoding/json"
	"ipc/helper"
)

type MetaBody struct {
	Type        MetaBodyType `json:"type"`
	SenderUID   uint64       `json:"sender_uid"`
	ReceiverUID uint64       `json:"receiver_uid"`
	Data        []byte       `json:"data"` // 注：json.Marshal时会把slice编码成base64
	StreamID    string       `json:"stream_id"`
	MetaID      string       `json:"meta_id"`
}

func NewMetaBody(senderUID uint64, data []byte, options ...MetaBodyOption) *MetaBody {
	mb := &MetaBody{SenderUID: senderUID, Data: data}

	for _, opt := range options {
		opt(mb)
	}

	if mb.MetaID == "" {
		randomBytes := make([]byte, 8)
		_, _ = rand.Read(randomBytes)
		mb.MetaID = helper.SimpleDecoder(randomBytes, "base64")
	}

	return mb
}

func (m *MetaBody) typeEncoding() DataEncoding {
	encoding := DataEncoding(m.Type & 0b11111110)
	switch encoding {
	case UTF8:
		return UTF8
	case BASE64:
		return BASE64
	case BINARY:
		return BINARY
	default:
		return UTF8
	}
}

func (m *MetaBody) typeIsInline() bool {
	return m.Type&INLINE != 0
}

func (m *MetaBody) typeIsStream() bool {
	return m.Type&INLINE == 0
}

func (m *MetaBody) MarshalJSON() ([]byte, error) {
	if m.typeEncoding() == BINARY {
		// Create an alias to avoid infinite recursion
		type MetaBodyAlias MetaBody

		metaBody := FromMetaBodyBase64(m.SenderUID, m.Data, WithStreamID(m.StreamID), WithReceiverUID(m.ReceiverUID))
		return json.Marshal((*MetaBodyAlias)(metaBody))
	}

	mb := struct {
		Type        MetaBodyType `json:"type"`
		SenderUID   uint64       `json:"sender_uid"`
		ReceiverUID uint64       `json:"receiver_uid"`
		Data        []byte       `json:"data"` // 注：json.Marshal时会把slice编码成base64
		StreamID    string       `json:"stream_id"`
		MetaID      string       `json:"meta_id"`
	}{m.Type, m.SenderUID, m.ReceiverUID, m.Data, m.StreamID, m.MetaID}

	return json.Marshal(mb)
}

type MetaBodyOption func(mb *MetaBody)

func WithMetaBodyType(mbType MetaBodyType) MetaBodyOption {
	return func(mb *MetaBody) {
		mb.Type = mbType
	}
}

func WithMetaID(metaID string) MetaBodyOption {
	return func(mb *MetaBody) {
		mb.MetaID = metaID
	}
}

func WithStreamID(streamID string) MetaBodyOption {
	return func(mb *MetaBody) {
		mb.StreamID = streamID
	}
}

func WithReceiverUID(receiverUID uint64) MetaBodyOption {
	return func(mb *MetaBody) {
		mb.ReceiverUID = receiverUID
	}
}

func FromMetaBodyText(senderUID uint64, data []byte, options ...MetaBodyOption) *MetaBody {
	mb := NewMetaBody(senderUID, data, options...)

	mb.Type = STREAM_WITH_TEXT
	if mb.StreamID == "" {
		mb.Type = INLINE_TEXT
	}

	return mb
}

func FromMetaBodyBase64(senderUID uint64, data []byte, options ...MetaBodyOption) *MetaBody {
	mb := NewMetaBody(senderUID, data, options...)
	mb.Type = STREAM_WITH_BASE64
	if mb.StreamID == "" {
		mb.Type = INLINE_BASE64
	}

	return mb
}

// FromMetaBodyBinary sender类型int64 | IPC
func FromMetaBodyBinary(sender interface{}, data []byte, options ...MetaBodyOption) *MetaBody {
	var mb *MetaBody

	switch v := sender.(type) {
	case uint64:
		mb = NewMetaBody(v, data, options...)
		mb.Type = STREAM_WITH_BINARY
		if mb.StreamID == "" {
			mb.Type = INLINE_BINARY
		}
	case IPC:
		if v.GetSupportBinary() {
			mb = FromMetaBodyBinary(v.GetUID(), data, options...)
		} else {
			// []byte在json.Marshal时，会把[]byte编码成base64，所以这里不需要提前
			// 使用helper.SimpleDecoder(data, "base64")进行base64编码
			// 在数据出口postMessage时，会进行json.Marshal
			mb = FromMetaBodyBase64(v.GetUID(), data, options...)
		}
	default:
		panic("metabody sender supports only int or IPC")
	}

	return mb
}

type MetaBodyType int

const (
	// STREAM_ID 流
	STREAM_ID MetaBodyType = 0

	// INLINE 内联数据
	INLINE MetaBodyType = 1

	// STREAM_WITH_TEXT 流，但是携带一帧的 UTF8 数据
	STREAM_WITH_TEXT MetaBodyType = STREAM_ID | MetaBodyType(UTF8)

	// STREAM_WITH_BASE64 流，但是携带一帧的 BASE64 数据
	STREAM_WITH_BASE64 MetaBodyType = STREAM_ID | MetaBodyType(BASE64)

	// STREAM_WITH_BINARY 流，但是携带一帧的 BINARY 数据
	STREAM_WITH_BINARY MetaBodyType = STREAM_ID | MetaBodyType(BINARY)

	// INLINE_TEXT 内联 UTF8 数据
	INLINE_TEXT MetaBodyType = INLINE | MetaBodyType(UTF8)

	// INLINE_BASE64 内联 BASE64 数据
	INLINE_BASE64 MetaBodyType = INLINE | MetaBodyType(BASE64)

	// INLINE_BINARY 内联 BINARY 数据
	INLINE_BINARY MetaBodyType = INLINE | MetaBodyType(BINARY)
)
