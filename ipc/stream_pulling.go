package ipc

import "fmt"

type StreamPulling struct {
	Type      MessageType
	StreamID  string
	Bandwidth *int
}

// NewStreamPulling
// 带宽限制, 0 代表不限速。
// 负数代表暂停，但对于数据暂停，一般使用 Paused 指令。
// 如果出现负数，往往是代表对方的数据处理出现了阻塞，与 Paused 不同，Paused 代表的是逻辑上的暂停，可能是挂起去处理其它事情去了，
// 而负数的带宽代表物理意义上的阻塞，此时更不该再发送更多的数据过去
func NewStreamPulling(streamID string, bandwidth *int) *StreamPulling {
	sp := &StreamPulling{
		Type:     STREAM_PULLING,
		StreamID: streamID,
	}

	if bandwidth == nil {
		sp.Bandwidth = new(int)
	} else {
		sp.Bandwidth = bandwidth
	}
	return sp
}

func (s *StreamPulling) String() string {
	return fmt.Sprintf("StreamPulling - Type: %d, StreamID: %s, Bandwidth: %p", s.Type, s.StreamID, s.Bandwidth)
}
