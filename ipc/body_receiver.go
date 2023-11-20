package ipc

import (
	"context"
	"fmt"
)

type BodyReceiver struct {
	Body
}

func NewBodyReceiver(metaBody *MetaBody, ipc IPC) *BodyReceiver {
	bRcv := &BodyReceiver{}

	if metaBody.typeIsStream() {
		if metaBody.StreamID == "" {
			panic("MetaBody stream id is empty")
		}
		metaID := fmt.Sprintf("%d/%s", metaBody.SenderUID, metaBody.StreamID)
		if _, ok := metaIDReceiverIpcMap[metaID]; !ok {
			ipc.OnClose(func(req any, ipc IPC) {
				delete(metaIDReceiverIpcMap, metaID)
			})
			metaIDReceiverIpcMap[metaID] = ipc
			metaBody.ReceiverUID = ipc.GetUID()
		}

		receiver, ok := metaIDReceiverIpcMap[metaID]
		if !ok {
			panic(fmt.Sprintf("no found ipc by metaID: %s", metaID))
		}
		ipc = receiver
		bRcv.bodyHub = NewBodyHub(metaToStream(metaBody, ipc))
	} else {
		switch metaBody.typeEncoding() {
		case UTF8:
			// 文本模式，直接返回即可，因为 RequestInit/Response 支持支持传入 utf8 字符串
			bRcv.bodyHub = NewBodyHub(string(metaBody.Data))
		case BASE64, BINARY:
			//data, err := helper.SimpleEncoder(string(metaBody.Data), "base64")
			//if err != nil {
			//	panic(fmt.Errorf("decode base64 err: %w", err))
			//}
			bRcv.bodyHub = NewBodyHub(metaBody.Data)
		default:
			panic(fmt.Sprintf("invalid metaBody type: %d", metaBody.Type))
		}
	}

	bRcv.metaBody = metaBody
	return bRcv
}

// FromBodyReceiver 返回值 *BodySender | *BodyReceiver
func FromBodyReceiver(metaBody *MetaBody, ipc IPC) BodyInter {
	if br, ok := metaIDIpcBodySenderMap[metaBody.MetaID]; ok {
		return br
	} else {
		return NewBodyReceiver(metaBody, ipc)
	}
}

func metaToStream(metaBody *MetaBody, ipc IPC) *ReadableStream {
	streamIPC := ipc
	if streamIPC == nil {
		panic("miss ipc when ipc-response has stream-body")
	}

	streamID := metaBody.StreamID
	paused := true

	onStart := func(ctrl *ReadableStreamDefaultController) {
		// ipc 一旦关闭，这个流也要关闭，因为只有这个ipc能为它提供数据
		streamIPC.OnClose(func(data any, ipc IPC) {
			ctrl.Close()
		})

		firstData := DataToBinary(metaBody.Data, metaBody.typeEncoding())
		if firstData != nil {
			_ = ctrl.Enqueue(firstData)
		}

		var unListen Observer
		unListen = streamIPC.OnStream(func(data any, ipc IPC) {
			if streamMsg, ok := IsStream(data); ok && streamMsg.StreamID == streamID {
				switch streamMsg.Type {
				case STREAM_DATA:
					if streamMsg.Encoding == BASE64 {
						_ = ctrl.Enqueue(streamMsg.Data)
					} else if d := DataToBinary(string(streamMsg.Data), streamMsg.Encoding); d != nil {
						_ = ctrl.Enqueue(d)
					}
				case STREAM_END:
					ctrl.Close()
					unListen(nil, nil)
				}
			}
		})
	}

	onPull := func(ctrl *ReadableStreamDefaultController) {
		if paused {
			paused = false
			_ = streamIPC.PostMessage(context.Background(), NewStreamPulling(streamID, nil))
		}
	}

	onCancel := func() {
		_ = streamIPC.PostMessage(context.Background(), NewStreamAbort(streamID))
	}

	// HighWaterMark=0，防止立即拉取stream
	return NewReadableStream(
		WithOnStart(onStart),
		WithOnPull(onPull),
		WithOnCancel(onCancel),
		WithHighWaterMark(0))
}
