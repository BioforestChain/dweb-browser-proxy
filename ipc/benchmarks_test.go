package ipc

import (
	"testing"
)

func BenchmarkIPC(b *testing.B) {
	readableStreamIpc := NewReadableStreamIPCWithDefaultInputStream(CLIENT, SupportProtocol{Raw: true})

	req := request()
	for i := 0; i < b.N; i++ {
		_ = readableStreamIpc.Enqueue(req)
	}
}
