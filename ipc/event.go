package ipc

import "fmt"

type Event struct {
	Type     MessageType  `json:"type"`
	Name     string       `json:"name"`
	Data     []byte       `json:"data"`
	Encoding DataEncoding `json:"encoding"`
}

func NewEvent(name string, data []byte, encoding DataEncoding) *Event {
	return &Event{Type: EVENT, Name: name, Data: data, Encoding: encoding}
}

func (s *Event) String() string {
	return fmt.Sprintf("Event - Type: %d, Name: %s, Data: %p, Encoding: %d", s.Type, s.Name, s.Data, s.Encoding)
}
