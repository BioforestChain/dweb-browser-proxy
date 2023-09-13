package ipc

import (
	"testing"
)

// TODO complete full process bench
func BenchmarkIPC(b *testing.B) {
	readableStreamIpc := NewReadableStreamIPCWithDefaultInputStream(CLIENT, SupportProtocol{Raw: true})

	req := request()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = readableStreamIpc.Enqueue(req)
	}
}
