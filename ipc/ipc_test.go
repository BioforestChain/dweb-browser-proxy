package ipc

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestBaseIPC_AllocReqID(t *testing.T) {
	ipc := NewBaseIPC()
	ipc.AllocReqID()
	reqID := ipc.AllocReqID()

	if reqID != 2 {
		t.Fatal("ipc req id inc failed")
	}
}

func TestBaseIPC_Request(t *testing.T) {
	t.Run("request with timeout", func(t *testing.T) {
		ipc := NewBaseIPC()
		req := ipc.Request("http://www.example.com", RequestArgs{Method: "GET"})

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		_, err := ipc.Send(ctx, req)
		if err == nil {
			t.Fatal("ipc request should timeout but not")
		}
	})

	t.Run("request", func(t *testing.T) {
		ipc := NewBaseIPC()

		go func() {
			ipc.msgSignal.Emit(NewResponse(1, 200, NewHeader(), nil, ipc), nil)
		}()

		req := ipc.Request("http://www.example.com", RequestArgs{Method: "GET"})
		res, err := ipc.Send(context.TODO(), req)
		if err != nil {
			t.Fatal("ipc request failed")
		}

		if res.Type != RESPONSE {
			t.Fatal("ipc request failed")
		}
	})

}

func TestBaseIPC_ConcurrentRequest(t *testing.T) {
	ipc := NewBaseIPC()

	type reqRes struct {
		req    *Request
		res    *Response
		resErr error
	}

	var mutex sync.Mutex
	var wg sync.WaitGroup

	testReqRes := make([]reqRes, 0)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req := ipc.Request("http://www.example.com", RequestArgs{Method: "GET"})
			res, err := ipc.Send(context.TODO(), req)

			mutex.Lock()
			defer mutex.Unlock()
			testReqRes = append(testReqRes, reqRes{req, res, err})
		}()
	}

	// emit response和send req并发的话，可能导致deadlock问题
	// 使用时一定是先有send req，后有emit response，所以这里加for等待req发送
	for len(ipc.reqResMap.getAll()) != 3 {
		time.Sleep(time.Millisecond * 10)
	}

	ipc.msgSignal.Emit(NewResponse(1, 200, NewHeader(), nil, ipc), nil)
	ipc.msgSignal.Emit(NewResponse(2, 200, NewHeader(), nil, ipc), nil)
	ipc.msgSignal.Emit(NewResponse(3, 200, NewHeader(), nil, ipc), nil)

	wg.Wait()

	if len(testReqRes) != 3 {
		t.Fatal("ipc concurrent request failed")
	}

	for _, rr := range testReqRes {
		if rr.resErr != nil {
			t.Fatal("ipc concurrent request failed: ", rr.resErr)
		}

		if rr.req.ID != rr.res.ReqID {
			t.Fatal("ipc concurrent request failed")
		}
	}
}

func TestBaseIPC_RequestTimeout(t *testing.T) {
	ipc := NewBaseIPC(WithReqTimeout(10 * time.Millisecond))

	req := ipc.Request("http://www.example.com", RequestArgs{Method: "GET"})
	_, err := ipc.Send(context.TODO(), req)
	if err == nil && !errors.Is(err, ErrReqTimeout) {
		t.Fatal("ipc request timeout failed")
	}
}

func TestBaseIPC_createSignal_Close(t *testing.T) {
	ipc := NewBaseIPC()
	var i int
	signal1 := ipc.createSignal(false)
	signal1.Listen(func(req any, ipc IPC) {
		i++
	})
	signal2 := ipc.createSignal(false)
	signal2.Listen(func(req any, ipc IPC) {
		i++
	})

	ipc.Close()

	signal1.Emit(nil, nil)
	signal2.Emit(nil, nil)

	if i != 0 {
		t.Fatal("ipc createSignal and Close failed")
	}
}
