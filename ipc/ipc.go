package ipc

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var ErrReqTimeout = errors.New("req timeout")

type IPC interface {
	Request(url string, init RequestArgs) *Request
	Send(ctx context.Context, req *Request) (*Response, error)
	PostMessage(ctx context.Context, msg interface{}) error
	GetUID() uint64
	GetSupportBinary() bool
	OnClose(observer Observer)
	Close()
	GetOutputStreamReader() *ReadableStreamDefaultReader
}

type BaseIPC struct {
	uid             uint64
	reqID           uint64
	supportBinary   bool
	SupportProtocol SupportProtocol

	msgSignal     *Signal
	requestSignal *Signal
	streamSignal  *Signal
	closeSignal   *Signal // 负责clear所有信号的observer
	closed        bool    // 所有信号是否被关闭

	reqResMap  *reqResMap
	mu         sync.Mutex
	reqTimeout time.Duration

	postMessage           func(ctx context.Context, msg interface{}) error // msg发送至outputStream
	doClose               func()                                           // inputStream被关闭时，需要close outputStream
	getOutputStreamReader func() *ReadableStreamDefaultReader              // 获取outputStream reader

	listenRequestOnce, listenResponseOnce, listenStreamOnce sync.Once
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
		ipc.postMessage = func(ctx context.Context, req interface{}) error {
			return nil
		}
	}

	ipc.closeSignal = NewSignal(false)
	ipc.msgSignal = ipc.createSignal(false)

	if ipc.reqTimeout == 0 {
		ipc.reqTimeout = 30 * time.Second
	}

	return ipc
}

func (bipc *BaseIPC) Request(url string, init RequestArgs) *Request {
	reqID := bipc.AllocReqID()
	return FromRequest(reqID, bipc, url, init)
}

// Send will wait for response
func (bipc *BaseIPC) Send(ctx context.Context, req *Request) (*Response, error) {
	//resCh := make(chan *Response)
	resCh := chanResponsePool.Get().(chan *Response)

	// register and listen response
	bipc.RegisterReqID(req.ID, resCh)

	defer func() {
		resChan, ok := bipc.reqResMap.getAndDelete(req.ID)
		if ok {
			chanResponsePool.Put(resChan)
		}
	}()

	// send req
	if err := bipc.PostMessage(ctx, req); err != nil {
		return nil, err
	}

	// wait response
	select {
	case res := <-resCh:
		return res, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(bipc.reqTimeout):
		return nil, fmt.Errorf("%w, id: %d\n", ErrReqTimeout, req.ID)
	}
}

func (bipc *BaseIPC) PostMessage(ctx context.Context, msg interface{}) error {
	return bipc.postMessage(ctx, msg)
}

func (bipc *BaseIPC) GetOutputStreamReader() *ReadableStreamDefaultReader {
	return bipc.getOutputStreamReader()
}

func (bipc *BaseIPC) OnMessage(observer Observer) {
	bipc.msgSignal.Listen(observer)
}

func (bipc *BaseIPC) OnRequest(observer Observer) {
	bipc.listenRequestOnce.Do(func() {
		bipc.requestSignal = func() *Signal {
			signal := bipc.createSignal(false)
			bipc.OnMessage(func(req interface{}, ipc IPC) {
				if _, ok := req.(*Request); !ok {
					return
				}
				signal.Emit(req, ipc)
			})
			return signal
		}()
	})
	bipc.requestSignal.Listen(observer)
}

func (bipc *BaseIPC) OnFetch() {
	// TODO 待实现
}

func (bipc *BaseIPC) OnStream(observer Observer) {
	bipc.listenStreamOnce.Do(func() {
		bipc.streamSignal = func() *Signal {
			signal := bipc.createSignal(false)
			bipc.OnMessage(func(req interface{}, ipc IPC) {
				// TODO 待实现
				//if reqStr, ok := req.(string); ok {
				//	if strings.Contains(reqStr, "stream_id") {
				//		signal.Emit(req, ipc)
				//	}
				//}
			})
			return signal
		}()
	})

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
	bipc.mu.Lock()
	defer bipc.mu.Unlock()

	bipc.reqID++
	return bipc.reqID
}

func (bipc *BaseIPC) RegisterReqID(reqID uint64, resCh chan *Response) {
	bipc.reqResMap.update(reqID, resCh)

	bipc.listenResponseOnce.Do(func() {
		bipc.OnMessage(func(oc interface{}, ipc IPC) {
			if res, ok := oc.(*Response); ok {
				if resChan, has := bipc.reqResMap.getAndDelete(res.ReqID); has {
					//bipc.reqResMap.delete(res.ReqID)
					resChan <- res
					chanResponsePool.Put(resChan)
					//close(resChan)
				}
			}
		})
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
	v  map[uint64]chan *Response
	mu sync.RWMutex
}

func newReqResMap() *reqResMap {
	return &reqResMap{v: make(map[uint64]chan *Response)}
}

func (r *reqResMap) get(reqID uint64) (resChan chan *Response, ok bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	resChan, ok = r.v[reqID]
	return resChan, ok
}

func (r *reqResMap) getAndDelete(reqID uint64) (resChan chan *Response, ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	resChan, ok = r.v[reqID]

	if ok {
		delete(r.v, reqID)
	}
	return resChan, ok
}

func (r *reqResMap) getAll() map[uint64]chan *Response {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.v
}

func (r *reqResMap) update(reqID uint64, resCh chan *Response) {
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

func WithPostMessage(postMsg func(ctx context.Context, req interface{}) error) Option {
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

func WithStreamReader(getOutputStreamReader func() *ReadableStreamDefaultReader) Option {
	return func(ipc *BaseIPC) {
		ipc.getOutputStreamReader = getOutputStreamReader
	}
}

type RequestArgs struct {
	Method string
	Body   interface{} // nil | "" | string | []byte | ReadableStream
	Header Header
}

var chanResponsePool = sync.Pool{New: func() any {
	return make(chan *Response)
}}
