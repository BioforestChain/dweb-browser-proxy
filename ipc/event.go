package ipc

import "fmt"

type Event struct {
	Type     MessageType
	Name     string
	Data     interface{}
	Encoding DataEncoding
}

func NewEvent(name string, data interface{}, encoding DataEncoding) *Event {
	return &Event{Type: EVENT, Name: name, Data: data, Encoding: encoding}
}

func (s *Event) String() string {
	return fmt.Sprintf("Event - Type: %d, Name: %s, Data: %p, Encoding: %d", s.Type, s.Name, s.Data, s.Encoding)
}
