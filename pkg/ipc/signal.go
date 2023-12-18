package ipc

import (
	"reflect"
	"sync"
)

type Signal struct {
	observers  []Observer
	started    bool
	cachedEmit []emitArgs
	mutex      sync.Mutex
}

type Observer func(data any, ipc IPC)

type emitArgs struct {
	data any
	ipc  IPC
}

func NewSignal(autoStart bool) *Signal {
	s := &Signal{}

	if autoStart {
		s.start()
	}
	return s
}

func (s *Signal) start() {
	if s.started {
		return
	}

	s.started = true
	if len(s.cachedEmit) > 0 {
		for _, args := range s.cachedEmit {
			s.emit(args.data, args.ipc)
		}
		s.cachedEmit = make([]emitArgs, 0)
	}
}

func (s *Signal) Listen(observer Observer) func() {
	s.mutex.Lock()
	s.observers = append(s.observers, observer)
	s.mutex.Unlock()

	s.start()
	return func() {
		s.UnListen(observer)
	}
}

func (s *Signal) Emit(data any, ipc IPC) {
	if s.started {
		s.emit(data, ipc)
	} else {
		s.mutex.Lock()
		s.cachedEmit = append(s.cachedEmit, emitArgs{data, ipc})
		s.mutex.Unlock()
	}
}

func (s *Signal) emit(data any, ipc IPC) {
	for _, ob := range s.observers {
		// TODO 改成异步的话，在body stream场景下，
		// 先后接收stream data和 stream end时，由于异步，可能先stream end，
		// 导致stream data接收不全
		// 同步处理的话，ob存在阻塞操作会导致后续ob无法执行
		ob(data, ipc)
		//go func(observer Observer) {
		//	observer(data, ipc)
		//}(ob)
	}
}

func (s *Signal) Clear() {
	s.observers = make([]Observer, 0)
}

func (s *Signal) UnListen(observer Observer) {
	for i, o := range s.observers {
		if reflect.ValueOf(o).Pointer() == reflect.ValueOf(observer).Pointer() {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			break
		}
	}
}

func (s *Signal) GetObservers() []Observer {
	return s.observers
}
