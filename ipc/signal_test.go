package ipc

import (
	"sync"
	"testing"
)

func TestSignalAutoStart(t *testing.T) {
	signal := NewSignal(true)

	signal.Listen(func(oc any, ipc IPC) {
		req, _ := oc.(*Request)
		if req.Type != REQUEST {
			t.Fatal("callback failed")
		}
	})

	ipc := NewBaseIPC()
	req := NewRequest(1, "http://www.exmaple.com", GET, Header{}, nil, ipc)
	signal.Emit(req, ipc)
}

func TestSignalLazyStart(t *testing.T) {
	signal := NewSignal(false)

	ipc := NewBaseIPC()
	req := NewRequest(1, "http://www.exmaple.com", GET, Header{}, nil, ipc)

	signal.Emit(req, ipc)
	signal.Emit(req, ipc)

	var i int

	var wg sync.WaitGroup
	wg.Add(2)

	// observer will be executed twice because there are two emits.
	signal.Listen(func(oc any, ipc IPC) {
		defer wg.Done()

		req, _ := oc.(*Request)
		i++
		if req.Type != REQUEST {
			t.Fatal("callback failed")
		}
	})
	wg.Wait()

	if i != 2 {
		t.Fatal("call times incorrect")
	}
}

func TestSignalMultiListen(t *testing.T) {
	signal := NewSignal(true)

	var wg sync.WaitGroup
	wg.Add(2)

	var i int
	signal.Listen(func(oc any, ipc IPC) {
		defer wg.Done()

		req, _ := oc.(*Request)
		i++
		if req.Type != REQUEST {
			t.Fatal("callback failed")
		}
	})

	signal.Listen(func(oc any, ipc IPC) {
		defer wg.Done()

		req, _ := oc.(*Request)
		i++
		if req.Type != REQUEST {
			t.Fatal("callback failed")
		}
	})

	ipc := NewBaseIPC()
	req := NewRequest(1, "http://www.exmaple.com", GET, Header{}, nil, ipc)
	signal.Emit(req, ipc)

	wg.Wait()
	if i != 2 {
		t.Fatal("multi listen op failed")
	}
}

func TestSignal_Unregister(t *testing.T) {
	signal := NewSignal(false)

	ipc := NewBaseIPC()
	req := NewRequest(1, "http://www.exmaple.com", GET, Header{}, nil, ipc)

	signal.Emit(req, ipc)

	var i int
	var wg sync.WaitGroup
	wg.Add(2)

	// observer will be executed twice because there are two emits.
	signal.Listen(func(oc any, ipc IPC) {
		defer wg.Done()

		req, _ := oc.(*Request)
		i++
		if req.Type != REQUEST {
			t.Fatal("callback failed")
		}
	})

	unRegister := signal.Listen(func(oc any, ipc IPC) {
		t.Fatal("callback failed")
	})
	unRegister()

	ipc1 := NewBaseIPC()
	req1 := NewRequest(1, "http://www.exmaple.com", GET, Header{}, nil, ipc)
	signal.Emit(req1, ipc1)

	wg.Wait()

	if i != 2 {
		t.Fatal("unregister observer failed")
	}
}

func TestSignal_Clear(t *testing.T) {
	signal := NewSignal(true)

	signal.Listen(func(oc any, ipc IPC) {})
	signal.Clear()

	obs := signal.GetObservers()

	if len(obs) != 0 {
		t.Fatal("callback failed")
	}
}
