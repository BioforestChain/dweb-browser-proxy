package ipc

import (
	"reflect"
	"testing"
)

func TestHeader(t *testing.T) {
	wanted := map[string]string{"access-control-allow-origin": "john"}

	h := NewHeader()
	h.init("access-control-allow-origin", wanted["access-control-allow-origin"])

	if !reflect.DeepEqual(h, Header(wanted)) {
		t.Fatal("new ipc header failed")
	}
}

func TestHeader_toJSON(t *testing.T) {
	wanted := map[string]string{"Access-Control-Allow-Origin": "john"}

	h := NewHeader()
	h.init("access-control-allow-origin", "john")
	newHeader := h.toJSON()

	if !reflect.DeepEqual(newHeader, wanted) {
		t.Fatal("new ipc header toJSON failed")
	}
}
