package helper

import (
	"bytes"
	"encoding/binary"
)

func U8aConcat(a []byte, b []byte) []byte {
	var buf bytes.Buffer
	buf.Write(a)
	buf.Write(b)
	return buf.Bytes()
}

// U32To8Concat 根据count生成4字节slice，并和b合并
// 用于格式化数据为：前4字节表示消息大小，后续字节表示消息
func U32To8Concat(count int, b []byte) []byte {
	n := make([]byte, 4)
	binary.LittleEndian.PutUint32(n, uint32(count))
	return U8aConcat(n, b)
}

func U8aToU32(a []byte) uint32 {
	return binary.LittleEndian.Uint32(a)
}

func BytesEqual(a []byte, b string) bool {
	return bytes.Equal(a, []byte(b))
}
