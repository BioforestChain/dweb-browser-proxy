package ipc

import (
	"encoding/json"
	"strings"
)

type Header map[string]string

func NewHeader() Header {
	return make(map[string]string)
}

func NewHeaderWithExtra(headers map[string]string) Header {
	if headers != nil {
		return headers
	}

	return NewHeader()
}

// init new Header().init("Content-Type", "text/html,charset=utf8"),
func (h Header) init(key string, value string) {
	if _, ok := h[key]; !ok {
		h[key] = value
	}
}

func (h Header) MarshalJSON() ([]byte, error) {
	newHeader := h.toJSON()
	return json.Marshal(newHeader)
}

func (h Header) Get(key string) string {
	return h[key]
}

// headers.toJSON()
// NewIpcHeader.toJSON
func (h Header) toJSON() map[string]string {
	newRecord := make(map[string]string)
	for k := range h {
		var newKey string
		if strings.Contains(k, "-") {
			parts := strings.Split(k, "-")
			for i, part := range parts {
				if i == 0 {
					newKey += strings.Title(part)
				} else {
					newKey += "-" + strings.Title(part)
				}
			}
		} else {
			newKey = strings.Title(k)
		}
		newRecord[newKey] = h[k]
	}
	return newRecord
}
