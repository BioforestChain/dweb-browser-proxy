package ipc

import (
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
		if _, ok := metaId_receiverIpc_Map[metaID]; !ok {
			ipc.OnClose(func(req interface{}, ipc IPC) {
				delete(metaId_receiverIpc_Map, metaID)
			})
			metaId_receiverIpc_Map[metaID] = ipc
			metaBody.ReceiverUID = ipc.GetUID()
		}

		receiver, ok := metaId_receiverIpc_Map[metaID]
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
	if br, ok := metaId_ipcBodySender_Map[metaBody.MetaID]; ok {
		return br
	} else {
		return NewBodyReceiver(metaBody, ipc)
	}
}

func metaToStream(metaBody *MetaBody, ipc IPC) *ReadableStream {
	// TODO 待实现
	return NewReadableStream()
}
