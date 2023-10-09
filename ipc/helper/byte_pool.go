package helper

import (
	"sync"
)

var (
	buf32kPool     sync.Pool
	buf16kPool     sync.Pool
	buf8kPool      sync.Pool
	buf2kPool      sync.Pool
	buf1kPool      sync.Pool
	bufPoolDefault sync.Pool
	headerPool     = newPool(4)
)

func newPool(size int) sync.Pool {
	return sync.Pool{
		New: func() any {
			return make([]byte, size)
		},
	}
}

func GetBuffer(size int) []byte {
	var buf any

	if size >= 32*1024 {
		buf = buf32kPool.Get()
	} else if size >= 16*1024 {
		buf = buf16kPool.Get()
	} else if size >= 8*1024 {
		buf = buf8kPool.Get()
	} else if size >= 2*1024 {
		buf = buf2kPool.Get()
	} else if size >= 1*1024 {
		buf = buf1kPool.Get()
	} else if size == 4 {
		buf = headerPool.Get()
	} else {
		buf = bufPoolDefault.Get()
	}

	if buf == nil {
		return make([]byte, size)
	}

	b := buf.([]byte)
	if cap(b) < size {
		return make([]byte, size)
	}

	return b[:size]
}

func PutBuffer(buf []byte) {
	var size int = cap(buf)

	if size > 64<<10 {
		return
	}

	buf = buf[:0]

	if size >= 32*1024 {
		buf32kPool.Put(buf)
	} else if size >= 16*1024 {
		buf16kPool.Put(buf)
	} else if size >= 8*1024 {
		buf8kPool.Put(buf)
	} else if size >= 2*1024 {
		buf2kPool.Put(buf)
	} else if size >= 1*1024 {
		buf1kPool.Put(buf)
	} else if size == 4 {
		headerPool.Put(buf)
	} else {
		bufPoolDefault.Put(buf)
	}
}
