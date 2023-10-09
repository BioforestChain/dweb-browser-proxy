package ipc

import (
	"context"
	"io"
	"sync"
)

type BodySender struct {
	Body
	isStream                       bool
	ipc                            IPC
	openSignal, closeSignal        *Signal
	controller                     *streamController
	isStreamOpened, isStreamClosed bool
	usedIpcMap                     map[IPC]*usedIpcInfo
}

// NewBodySender data类型包括：string | []byte | *ReadableStream
func NewBodySender(data any, ipc IPC) *BodySender {
	bs := &BodySender{
		ipc:         ipc,
		openSignal:  NewSignal(false),
		closeSignal: NewSignal(false),
		usedIpcMap:  make(map[IPC]*usedIpcInfo),
	}
	bs.bodyHub = NewBodyHub(data)
	bs.metaBody = bs.bodyAsMeta(data)
	if bs.bodyHub.Stream != nil {
		bs.isStream = true

		rawIpcBodyWMap[data.(*ReadableStream)] = bs
	}

	// 作为 "生产者"，第一持有这个 IpcBodySender
	UsableByIpc(ipc, bs)
	return bs
}

func (bs *BodySender) onStreamOpen(observer Observer) {
	bs.openSignal.Listen(observer)
}

func (bs *BodySender) onStreamClose(observer Observer) {
	bs.closeSignal.Listen(observer)
}

// emitStreamPull 拉取数据
func (bs *BodySender) emitStreamPull(info *usedIpcInfo, msg *StreamPulling) {
	info.bandwidth = *msg.Bandwidth
	bs.controller.Pulling()
}

// emitStreamPaused 暂停数据
func (bs *BodySender) emitStreamPaused(info *usedIpcInfo, msg *StreamPaused) {
	// 更新保险限制
	info.bandwidth = -1
	info.fuse = *msg.fuse

	// 如果所有的读取者都暂停了，那么就触发暂停
	paused := true
	for _, v := range bs.usedIpcMap {
		if v.bandwidth >= 0 {
			paused = false
			break
		}
	}

	if paused {
		bs.controller.Paused()
	}
}

func (bs *BodySender) emitStreamAborted(info *usedIpcInfo) {
	if v, ok := bs.usedIpcMap[info.ipc]; ok && v != nil {
		delete(bs.usedIpcMap, info.ipc)
		// 如果没有任何消费者了，那么真正意义上触发 abort
		if len(bs.usedIpcMap) == 0 {
			bs.controller.Aborted()
		}
	}
}

func (bs *BodySender) emitStreamClose() {
	bs.setIsStreamOpened(true)
	bs.setIsStreamClosed(true)
}

func (bs *BodySender) setIsStreamOpened(v bool) {
	if bs.isStreamOpened != v {
		bs.isStreamOpened = v
		if v {
			bs.openSignal.Emit(nil, nil)
			bs.openSignal.Clear()
		}
	}
}

func (bs *BodySender) setIsStreamClosed(v bool) {
	if bs.isStreamClosed != v {
		bs.isStreamClosed = v
		if v {
			bs.closeSignal.Emit(nil, nil)
			bs.closeSignal.Clear()
		}
	}
}

func (bs *BodySender) bodyAsMeta(body any) *MetaBody {
	switch v := body.(type) {
	case string:
		return FromMetaBodyText(bs.ipc.GetUID(), []byte(v))
	case []byte:
		return FromMetaBodyBinary(bs.ipc, v)
	case *ReadableStream:
		return bs.streamAsMeta(v)
	default:
		panic("bodyAsMeta body supports only string or ReadableStream or []byte")
	}
}

// 如果 rawData 是流模式，需要提供数据发送服务
// 这里不会一直无脑发，而是对方有需要的时候才发
func (bs *BodySender) streamAsMeta(stream *ReadableStream) *MetaBody {
	var streamFirstData []byte

	streamID := GetStreamID(stream)

	// 读取流数据操作
	pullingFunc := func(streamCtrl *streamController) {
		data, err := streamCtrl.streamRead.readAny()
		if err == io.EOF {
			msg := NewStreamEnd(streamID)
			// 不论是不是被 aborted，都发送结束信号
			for ipc, _ := range bs.usedIpcMap {
				_ = ipc.PostMessage(context.Background(), msg)
			}

			// 退出流控制
			streamCtrl.Aborted()

			// 关闭流
			streamCtrl.streamRead.abortStream()
			bs.emitStreamClose()
			return
		}

		bs.setIsStreamOpened(true)

		msg := FromStreamDataBinary(streamID, data)
		for ipc, _ := range bs.usedIpcMap {
			_ = ipc.PostMessage(context.Background(), msg)
		}
	}

	// abort流操作
	abortFunc := func(controller *streamController) {
		// 关闭流
		controller.streamRead.abortStream()
		bs.emitStreamClose()
	}

	bs.controller = newStreamController(
		stream,
		WithStreamControllerPullingFunc(pullingFunc),
		WithStreamControllerAbortFunc(abortFunc))

	mb := NewMetaBody(
		bs.ipc.GetUID(),
		streamFirstData,
		WithMetaBodyType(STREAM_ID),
		WithStreamID(streamID),
	)

	metaIDIpcBodySenderMap[mb.MetaID] = bs
	bs.controller.Listen(func(streamSignal any, ipc IPC) {
		if _, ok := streamSignal.(streamCtorSignalAborted); ok {
			delete(metaIDIpcBodySenderMap, mb.MetaID)
		}
	})

	return mb
}

func (bs *BodySender) useByIpc(ipc IPC) (info *usedIpcInfo) {
	info, ok := bs.usedIpcMap[ipc]
	if info != nil && ok {
		return
	}

	if bs.isStream && !bs.isStreamOpened {
		info = newUsedIpcInfo(bs, ipc)
		bs.usedIpcMap[ipc] = info
		bs.onStreamClose(func(data any, ipc IPC) {
			bs.emitStreamAborted(info)
		})
		return
	}

	return
}

var UsableIpcBodyMap = make(map[IPC]*usableIpcBodyMapper)

// UsableByIpc 监听流及处理
func UsableByIpc(ipc IPC, body *BodySender) {
	if !(body.isStream && !body.isStreamOpened) {
		return
	}

	streamID := body.metaBody.StreamID

	var usableIpcBodyM *usableIpcBodyMapper
	var ok bool
	if usableIpcBodyM, ok = UsableIpcBodyMap[ipc]; !ok {
		usableIpcBodyM = newUsableIpcBodyMapper()
		usableIpcBodyM.onDestroy(ipc.OnStream(func(data any, ipc IPC) {
			t, ok := IsStream(data)
			if !ok {
				return
			}

			switch t.Type {
			case STREAM_PULLING:
				msg := data.(*StreamPulling)
				if bodySender := usableIpcBodyM.get(msg.StreamID); bodySender != nil {
					if ipcInfo := bodySender.useByIpc(ipc); ipcInfo != nil {
						ipcInfo.emitStreamPull(msg)
					}
				}
			case STREAM_PAUSED:
				msg := data.(*StreamPaused)
				if bodySender := usableIpcBodyM.get(msg.StreamID); bodySender != nil {
					if ipcInfo := bodySender.useByIpc(ipc); ipcInfo != nil {
						ipcInfo.emitStreamPaused(msg)
					}
				}
			case STREAM_ABORT:
				msg := data.(*StreamAbort)
				if bodySender := usableIpcBodyM.get(msg.StreamID); bodySender != nil {
					if ipcInfo := bodySender.useByIpc(ipc); ipcInfo != nil {
						ipcInfo.emitStreamAborted()
					}
				}
			}
		}))
		usableIpcBodyM.onDestroy(func(data any, none IPC) {
			delete(UsableIpcBodyMap, ipc)
		})

		UsableIpcBodyMap[ipc] = usableIpcBodyM
	}

	if usableIpcBodyM.add(streamID, body) {
		// 一个流一旦关闭，那么就将不再会与它有主动通讯上的可能
		body.onStreamClose(func(data any, ipc IPC) {
			usableIpcBodyM.remove(streamID)
		})
	}
}

func GetStreamID(stream *ReadableStream) string {
	return stream.ID
}

func FromBodySenderText(data string, ipc IPC) *BodySender {
	return NewBodySender(data, ipc)
}

func FromBodySenderBinary(data []byte, ipc IPC) *BodySender {
	return NewBodySender(data, ipc)
}

func FromBodySenderStream(data *ReadableStream, ipc IPC) *BodySender {
	bodySender, ok := rawIpcBodyWMap[data]
	if ok && bodySender != nil {
		return bodySender.(*BodySender)
	}

	return NewBodySender(data, ipc)
}

type usableIpcBodyMapper struct {
	destroySignal *Signal
	m             map[string]*BodySender
}

func newUsableIpcBodyMapper() *usableIpcBodyMapper {
	return &usableIpcBodyMapper{
		destroySignal: NewSignal(false),
		m:             make(map[string]*BodySender),
	}
}

func (u *usableIpcBodyMapper) add(streamID string, body *BodySender) bool {
	if _, ok := u.m[streamID]; ok {
		return true
	}

	u.m[streamID] = body
	return false
}

func (u *usableIpcBodyMapper) get(streamID string) *BodySender {
	return u.m[streamID]
}

func (u *usableIpcBodyMapper) remove(streamID string) {
	if body, ok := u.m[streamID]; ok && body != nil {
		delete(u.m, streamID)
		if len(u.m) == 0 {
			u.destroySignal.Emit(nil, nil)
			u.destroySignal.Clear()
		}
	}
}

func (u *usableIpcBodyMapper) onDestroy(observer Observer) {
	u.destroySignal.Listen(observer)
}

// usedIpcInfo 被哪些 ipc 所真正使用，使用的进度分别是多少
// 这个进度 用于 类似流的 多发
type usedIpcInfo struct {
	body            *BodySender
	ipc             IPC
	bandwidth, fuse int
}

func newUsedIpcInfo(body *BodySender, ipc IPC) *usedIpcInfo {
	return &usedIpcInfo{body: body, ipc: ipc}
}

func (u *usedIpcInfo) emitStreamPull(msg *StreamPulling) {
	u.body.emitStreamPull(u, msg)
}

func (u *usedIpcInfo) emitStreamPaused(msg *StreamPaused) {
	u.body.emitStreamPaused(u, msg)
}

func (u *usedIpcInfo) emitStreamAborted() {
	u.body.emitStreamAborted(u)
}

// streamController 控制流的读取
type streamController struct {
	streamRead                        *binaryStreamRead
	started, pulling, paused, aborted bool
	pullingFunc, pauseFunc, abortFunc func(controller *streamController)
	startCh, pullingCh                chan struct{}
	signal                            *Signal
	startOnce                         sync.Once
}

func newStreamController(stream *ReadableStream, opts ...StreamControllerOption) *streamController {
	sc := &streamController{
		streamRead: newBinaryStreamRead(stream),
		startCh:    make(chan struct{}),
		pullingCh:  make(chan struct{}, 1),
		signal:     NewSignal(false),
	}

	for _, opt := range opts {
		opt(sc)
	}

	go func() {
		unListen := sc.signal.Listen(func(streamSignal any, ipc IPC) {
			switch streamSignal.(type) {
			case streamCtorSignalPulling:
				sc.pulling, sc.paused = true, false
				if len(sc.pullingCh) == 0 {
					sc.pullingCh <- struct{}{}
				}
				sc.start()
			case streamCtorSignalPaused:
				sc.pulling, sc.paused = false, true
				if len(sc.pullingCh) == 1 {
					<-sc.pullingCh
				}
				sc.start()
			case streamCtorSignalAborted:
				sc.pulling, sc.paused, sc.aborted = false, false, true
				if len(sc.pullingCh) == 0 {
					sc.pullingCh <- struct{}{}
				}
				sc.start()

				//if sc.abortFunc != nil {
				//	sc.abortFunc(sc)
				//}
			}
		})

		defer func() {
			unListen()
		}()

		for {
			// 防止控制stream之前，cpu空转
			if !sc.started {
				select {
				case <-sc.startCh:
				}
			}

			if sc.aborted {
				if sc.abortFunc != nil {
					sc.abortFunc(sc)
				}
				return
			}

			if sc.paused {
				if sc.pauseFunc != nil {
					sc.pauseFunc(sc)
				}

				select {
				case <-sc.pullingCh:
				}
			}

			if sc.pulling {
				if sc.pullingFunc != nil {
					sc.pullingFunc(sc)
				}
			}
		}
	}()

	return sc
}

func (s *streamController) start() {
	s.startOnce.Do(func() {
		s.started = true
		s.startCh <- struct{}{}
	})
}

func (s *streamController) Listen(observer Observer) {
	s.signal.Listen(observer)
}

func (s *streamController) Pulling() {
	s.signal.Emit(streamCtorSignalPulling{}, nil)
}

func (s *streamController) Paused() {
	s.signal.Emit(streamCtorSignalPaused{}, nil)
}

func (s *streamController) Aborted() {
	s.pulling, s.paused, s.aborted = false, false, true
	s.signal.Emit(streamCtorSignalAborted{}, nil)
}

type streamCtorSignalPulling struct{}
type streamCtorSignalPaused struct{}
type streamCtorSignalAborted struct{}

type StreamControllerOption func(sc *streamController)

func WithStreamControllerPullingFunc(pullingFunc func(controller *streamController)) StreamControllerOption {
	return func(sc *streamController) {
		sc.pullingFunc = pullingFunc
	}
}

func WithStreamControllerPauseFunc(pauseFunc func(controller *streamController)) StreamControllerOption {
	return func(sc *streamController) {
		sc.pauseFunc = pauseFunc
	}
}

func WithStreamControllerAbortFunc(abortFunc func(controller *streamController)) StreamControllerOption {
	return func(sc *streamController) {
		sc.abortFunc = abortFunc
	}
}
