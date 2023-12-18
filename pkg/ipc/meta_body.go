package ipc

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/ipc/helper"
	"strconv"
)

type MetaBody struct {
	Type        MetaBodyType `json:"type"`
	SenderUID   uint64       `json:"senderUid"`
	ReceiverUID uint64       `json:"receiverUid"`
	Data        []byte       `json:"data"` // 注：json.Marshal时会把slice编码成base64
	StreamID    string       `json:"streamId"`
	MetaID      string       `json:"metaId"`
}

func NewMetaBody(mType MetaBodyType, senderUID uint64, data []byte, options ...MetaBodyOption) *MetaBody {
	mb := &MetaBody{Type: mType, SenderUID: senderUID, Data: data}

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

// MarshalJSON
// outbound或序列化时: metaBody的data为base64编码值，且metaBody的type为INLINE_BASE64 || STREAM_WITH_BASE64
func (m *MetaBody) MarshalJSON() ([]byte, error) {
	// metaBodyAlias Create an alias to avoid infinite recursion
	type metaBodyAlias MetaBody

	metaBody := FromMetaBodyBase64(m.SenderUID, m.Data,
		WithStreamID(m.StreamID),
		WithMetaID(m.MetaID),
		WithReceiverUID(m.ReceiverUID))

	var mba = (*metaBodyAlias)(metaBody)
	return json.Marshal(mba)
}

// UnmarshalJSON
// inbound或反序列化时，
// 1. metaBody的type为（INLINE_TEXT || STREAM_WITH_TEXT）时，metaBody的data值为string类型，赋值给MetaBody.Data时需转换为[]byte
// 2. metaBody的type为（INLINE_BASE64 || STREAM_WITH_BASE64）时，metaBody的data值为string类型， 赋值给MetaBody.Data时先decode base64，再转换成[]byte
// 3. metaBody的type为（INLINE_BINARY || STREAM_WITH_BINARY），metaBody的data值为[]byte类型，赋值给MetaBody.Data时无需转换
// 4. MetaBody.Type类型设置为（INLINE_BINARY || STREAM_WITH_BINARY）
func (m *MetaBody) UnmarshalJSON(d []byte) error {
	// metaBodyAlias Create an alias to avoid infinite recursion
	type metaBodyAlias MetaBody
	var mba = &struct {
		*metaBodyAlias
		Data any `json:"data"`
	}{
		metaBodyAlias: (*metaBodyAlias)(m),
	}

	if err := json.Unmarshal(d, &mba); err != nil {
		return err
	}

	encoding := DataEncoding(mba.Type & 0b11111110)
	switch encoding {
	case UTF8:
		m.Data = []byte(mba.Data.(string))
	case BASE64:
		var err error
		m.Data, err = decodeBase64(mba.Data.(string))
		if err != nil {
			return err
		}
	case BINARY:
		m.Data = mba.Data.([]byte)
	default:
		m.Data = []byte(mba.Data.(string))
		//return errors.New("unknown data encoding")
	}

	if m.StreamID == "" {
		m.Type = INLINE_BINARY
	} else {
		m.Type = STREAM_WITH_BINARY
	}

	return nil
}

type MetaBodyOption func(mb *MetaBody)

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
	mb := NewMetaBody(STREAM_WITH_TEXT, senderUID, data, options...)
	if mb.StreamID == "" {
		mb.Type = INLINE_TEXT
	}

	return mb
}

func FromMetaBodyBase64(senderUID uint64, data []byte, options ...MetaBodyOption) *MetaBody {
	mb := NewMetaBody(STREAM_WITH_BASE64, senderUID, data, options...)
	if mb.StreamID == "" {
		mb.Type = INLINE_BASE64
	}

	return mb
}

// FromMetaBodyBinary sender类型uint64 | IPC
func FromMetaBodyBinary(sender any, data []byte, options ...MetaBodyOption) *MetaBody {
	var mb *MetaBody

	switch v := sender.(type) {
	case uint64:
		mb = NewMetaBody(STREAM_WITH_BINARY, v, data, options...)
		if mb.StreamID == "" {
			mb.Type = INLINE_BINARY
		}
	case IPC:
		if v.GetSupportBinary() {
			mb = FromMetaBodyBinary(v.GetUID(), data, options...)
		} else {
			// []byte在json.Marshal时，会把[]byte编码成base64
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

	// STREAM_WITH_TEXT 流 (2)，但是携带一帧的 UTF8 数据
	STREAM_WITH_TEXT MetaBodyType = STREAM_ID | MetaBodyType(UTF8)

	// STREAM_WITH_BASE64 流 (4)，但是携带一帧的 BASE64 数据
	STREAM_WITH_BASE64 MetaBodyType = STREAM_ID | MetaBodyType(BASE64)

	// STREAM_WITH_BINARY 流 (8)，但是携带一帧的 BINARY 数据
	STREAM_WITH_BINARY MetaBodyType = STREAM_ID | MetaBodyType(BINARY)

	// INLINE_TEXT 内联 UTF8 数据 (3)
	INLINE_TEXT MetaBodyType = INLINE | MetaBodyType(UTF8)

	// INLINE_BASE64 内联 BASE64 数据 (5)
	INLINE_BASE64 MetaBodyType = INLINE | MetaBodyType(BASE64)

	// INLINE_BINARY 内联 BINARY 数据 (9)
	INLINE_BINARY MetaBodyType = INLINE | MetaBodyType(BINARY)
)

type MetaBodyData []byte

func (m *MetaBodyData) UnmarshalJSON(d []byte) error {
	v, err := strconv.Unquote(string(d))
	if err != nil {
		return err
	}

	if result, ok := decodeBase64Old(v); ok {
		*m = result
	} else {
		*m = []byte(v)
	}

	return nil
}

func decodeBase64Old(src string) ([]byte, bool) {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(src)))
	n, err := base64.StdEncoding.Decode(dst, []byte(src))

	return dst[:n], err == nil
}

func decodeBase64(src string) ([]byte, error) {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(src)))
	n, err := base64.StdEncoding.Decode(dst, []byte(src))

	return dst[:n], err
}
