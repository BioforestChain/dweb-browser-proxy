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

// u32To8Concat 由data长度生成4字节slice，并和data合并，
// 返回数据格式为：前缀4字节表示消息大小，后缀字节表示消息内容
func u32To8Concat(data []byte) []byte {
	prefix := make([]byte, 4)
	binary.LittleEndian.PutUint32(prefix, uint32(len(data)))
	return U8aConcat(prefix, data)
}

func U8aToU32(a []byte) int {
	return int(binary.LittleEndian.Uint32(a))
}

func BytesEqual(a []byte, b string) bool {
	return bytes.Equal(a, []byte(b))
}

func GetBodySize(header []byte) int {
	return U8aToU32(header)
}

// FormatIPCData ipc数据格式为：前缀4字节表示消息大小，后缀字节表示消息内容
func FormatIPCData(data []byte) []byte {
	return u32To8Concat(data)
}
