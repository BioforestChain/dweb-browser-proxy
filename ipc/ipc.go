package ipc

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var ErrReqTimeout = errors.New("req timeout")

type IPC interface {
	Request(url string, init RequestArgs) *Request
	Send(req *Request) (*Response, error)
	PostMessage(msg interface{}) error // msg入队输出流
	GetUID() uint64
	GetSupportBinary() bool
	OnClose(observer Observer)
	Close()
	GetStreamReader() *ReadableStreamDefaultReader // 获取输出流的channel
}

type BaseIPC struct {
	uid             uint64
	reqID           uint64
	supportBinary   bool
	SupportProtocol SupportProtocol
	msgSignal       *Signal
	requestSignal   *Signal
	streamSignal    *Signal
	closeSignal     *Signal // 负责clear所有信号的observer
	closed          bool    // 所有信号是否被关闭
	reqResMap       *reqResMap
	mutex           sync.Mutex
	reqTimeout      time.Duration
	postMessage     func(msg interface{}) error         // msg 入队ipc输出流
	doClose         func()                              // proxy stream被关闭时，需要close输出流
	getStreamReader func() *ReadableStreamDefaultReader // 获取ipc输出流read
}

var UID uint64

func NewBaseIPC(opts ...Option) *BaseIPC {
	atomic.AddUint64(&UID, 1)

	ipc := &BaseIPC{
		uid:       atomic.LoadUint64(&UID),
		reqResMap: newReqResMap(),
	}

	for _, opt := range opts {
		opt(ipc)
	}

	if ipc.postMessage == nil {
		ipc.postMessage = func(req interface{}) error {
			return nil
		}
	}

	ipc.closeSignal = NewSignal(false)
	ipc.msgSignal = ipc.createSignal(false)
	ipc.requestSignal = func() *Signal {
		signal := ipc.createSignal(false)
		ipc.OnMessage(func(req interface{}, ipc IPC) {
			if _, ok := req.(*Request); !ok {
				return
			}
			signal.Emit(req, ipc)
		})
		return signal
	}()
	ipc.streamSignal = func() *Signal {
		signal := ipc.createSignal(false)
		ipc.OnMessage(func(req interface{}, ipc IPC) {
			// TODO 待实现
			//if reqStr, ok := req.(string); ok {
			//	if strings.Contains(reqStr, "stream_id") {
			//		signal.Emit(req, ipc)
			//	}
			//}
		})
		return signal
	}()

	if ipc.reqTimeout == 0 {
		ipc.reqTimeout = 60 * time.Second
	}

	return ipc
}

func (bipc *BaseIPC) Request(url string, init RequestArgs) *Request {
	reqID := bipc.AllocReqID()
	return FromRequest(reqID, bipc, url, init)
}

// Send will wait for response
func (bipc *BaseIPC) Send(req *Request) (*Response, error) {
	resCh := make(chan *Response)
	// register and listen response
	bipc.RegisterReqID(req.ID, resCh)

	// send req
	err := bipc.postMessage(req)
	if err != nil {
		return nil, err
	}

	// wait response
	// TODO res被读取后，需要close channel并把reqResMap对应记录删除
	// 可以通过signal来实现记录删除，对应看ipc.ts的close()实现
	select {
	case res := <-resCh:
		return res, nil
	case <-time.After(bipc.reqTimeout):
		return nil, fmt.Errorf("%w, id: %d\n", ErrReqTimeout, req.ID)
	}
}

func (bipc *BaseIPC) PostMessage(msg interface{}) error {
	return bipc.postMessage(msg)
}

func (bipc *BaseIPC) GetStreamReader() *ReadableStreamDefaultReader {
	return bipc.getStreamReader()
}

func (bipc *BaseIPC) OnMessage(observer Observer) {
	bipc.msgSignal.Listen(observer)
}

func (bipc *BaseIPC) OnRequest(observer Observer) {
	bipc.requestSignal.Listen(observer)
}

func (bipc *BaseIPC) OnFetch() {
	// TODO 待实现
}

func (bipc *BaseIPC) OnStream(observer Observer) {
	bipc.streamSignal.Listen(observer)
}

func (bipc *BaseIPC) OnClose(observer Observer) {
	bipc.closeSignal.Listen(observer)
}

func (bipc *BaseIPC) GetUID() uint64 {
	return bipc.uid
}

func (bipc *BaseIPC) GetSupportBinary() bool {
	return bipc.supportBinary
}

func (bipc *BaseIPC) AllocReqID() uint64 {
	bipc.mutex.Lock()
	defer bipc.mutex.Unlock()

	bipc.reqID++
	return bipc.reqID
}

func (bipc *BaseIPC) RegisterReqID(reqID uint64, resCh chan *Response) {
	bipc.reqResMap.update(reqID, resCh)

	bipc.OnMessage(func(oc interface{}, ipc IPC) {
		if res, ok := oc.(*Response); ok && res.Type == RESPONSE {
			if resch, has := bipc.reqResMap.get(res.ReqID); has {
				bipc.reqResMap.delete(res.ReqID)
				resch <- res
				close(resch)
			}
		}
	})
}

func (bipc *BaseIPC) Close() {
	if bipc.closed {
		return
	}
	bipc.closed = true

	if bipc.doClose != nil {
		bipc.doClose()
	}

	bipc.closeSignal.Emit(nil, nil)
	bipc.closeSignal.Clear()
}

func (bipc *BaseIPC) createSignal(autoStart bool) *Signal {
	signal := NewSignal(autoStart)
	bipc.OnClose(func(req interface{}, ipc IPC) {
		signal.Clear()
	})
	return signal
}

type reqResMap struct {
	v  map[uint64]chan<- *Response
	mu sync.Mutex
}

func newReqResMap() *reqResMap {
	return &reqResMap{v: make(map[uint64]chan<- *Response)}
}

func (r *reqResMap) get(reqID uint64) (resChan chan<- *Response, ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	resChan, ok = r.v[reqID]
	return resChan, ok
}

func (r *reqResMap) getAll() map[uint64]chan<- *Response {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.v
}

func (r *reqResMap) update(reqID uint64, resCh chan<- *Response) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.v[reqID] = resCh
}

func (r *reqResMap) delete(reqID uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.v, reqID)
}

type Option func(ipc *BaseIPC)

func WithReqTimeout(duration time.Duration) Option {
	return func(ipc *BaseIPC) {
		ipc.reqTimeout = duration
	}
}

func WithPostMessage(postMsg func(req interface{}) error) Option {
	return func(ipc *BaseIPC) {
		ipc.postMessage = postMsg
	}
}

func WithSupportProtocol(protocol SupportProtocol) Option {
	return func(ipc *BaseIPC) {
		ipc.SupportProtocol = protocol

		if protocol.MessagePack || protocol.ProtoBuf || protocol.Raw {
			ipc.supportBinary = true
		}
	}
}

func WithDoClose(doClose func()) Option {
	return func(ipc *BaseIPC) {
		ipc.doClose = doClose
	}
}

func WithStreamReader(getStreamReader func() *ReadableStreamDefaultReader) Option {
	return func(ipc *BaseIPC) {
		ipc.getStreamReader = getStreamReader
	}
}

type RequestArgs struct {
	Method string
	Body   interface{} // nil | "" | string | []byte | ReadableStream
	Header Header
}
