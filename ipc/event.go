package ipc

type Event struct {
	Type     MessageType
	Name     string
	Data     interface{}
	Encoding DataEncoding
}

func NewEvent(name string, data interface{}, encoding DataEncoding) *Event {
	return &Event{Type: EVENT, Name: name, Data: data, Encoding: encoding}
}
