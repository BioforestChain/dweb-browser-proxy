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

type Observer func(req interface{}, ipc IPC)

type emitArgs struct {
	req interface{}
	ipc IPC
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
			s.emit(args.req, args.ipc)
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

func (s *Signal) Emit(req interface{}, ipc IPC) {
	if s.started {
		s.emit(req, ipc)
	} else {
		s.mutex.Lock()
		s.cachedEmit = append(s.cachedEmit, emitArgs{req, ipc})
		s.mutex.Unlock()
	}
}

func (s *Signal) emit(req interface{}, ipc IPC) {
	for _, ob := range s.observers {
		//ob(req, ipc)
		go func(observer Observer) {
			observer(req, ipc)
		}(ob)
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
