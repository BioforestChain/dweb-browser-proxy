package ipc

import "strings"

type METHOD string

const (
	GET     METHOD = "GET"
	POST    METHOD = "POST"
	PUT     METHOD = "PUT"
	DELETE  METHOD = "DELETE"
	OPTIONS METHOD = "OPTIONS"
	TRACE   METHOD = "TRACE"
	PATCH   METHOD = "PATCH"
	PURGE   METHOD = "PURGE"
	HEAD    METHOD = "HEAD"
)

func ToIPCMethod(method string) METHOD {
	switch strings.ToUpper(method) {
	case string(GET):
		return GET
	case string(POST):
		return POST
	case string(PUT):
		return PUT
	case string(DELETE):
		return DELETE
	case string(OPTIONS):
		return OPTIONS
	case string(TRACE):
		return TRACE
	case string(PATCH):
		return PATCH
	case string(PURGE):
		return PURGE
	case string(HEAD):
		return HEAD
	default:
		return GET
	}
}

type ROLE string

const (
	SERVER ROLE = "server"
	CLIENT ROLE = "client"
)

type MessageType int

const (
	// END 特殊位：结束符
	// END = 1

	// REQUEST 类型：请求
	REQUEST MessageType = iota + 1

	// RESPONSE 类型：响应
	RESPONSE

	// STREAM_DATA 类型：流数据，发送方
	STREAM_DATA

	// STREAM_PULLING 类型：流拉取，请求方
	// 发送方一旦收到该指令，就可以持续发送数据
	// 该指令中可以携带一些“限流协议信息”，如果违背该协议，请求方可能会断开连接
	STREAM_PULLING

	// STREAM_PAUSED 类型：流暂停，请求方
	// 发送方一旦收到该指令，就应当停止基本的数据发送
	// 该指令中可以携带一些“保险协议信息”，描述仍然允许发送的一些数据类型、发送频率等。如果违背该协议，请求方可以会断开连接
	STREAM_PAUSED

	// STREAM_END 类型：流关闭，发送方
	// 可能是发送完成了，也有可能是被中断了
	STREAM_END

	// STREAM_ABORT 类型：流中断，请求方
	STREAM_ABORT

	// EVENT 类型：事件
	EVENT
)

// DataEncoding 数据编码格式
type DataEncoding int

const (
	// UTF8 文本 json html 等
	UTF8 DataEncoding = 1 << 1
	// BASE64 使用文本表示的二进制
	BASE64 DataEncoding = 1 << 2
	// BINARY 二进制
	BINARY DataEncoding = 1 << 3
)

type GMessage interface {
	*Request | Response | Event | GStreamMessage
}

type GStreamMessage interface {
	StreamData | StreamPulling | StreamPaused | StreamAbort | StreamEnd
}

type GBodyHub interface {
	ReadableStream | string | []byte
}

type GBody interface {
	BodySender | BodyReceiver
}
