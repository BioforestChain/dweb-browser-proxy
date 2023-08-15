package ipc

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// ReadableStream Streams API的不完全实现 https://developer.mozilla.org/en-US/docs/Web/API/Streams_API
// controller.close功能不完全实现
// backpressure/queuing strategies机制未实现
// reader.cancel功能未实现
type ReadableStream struct {
	ID            string
	dataChan      chan []byte
	mutex         sync.RWMutex
	highWaterMark uint64 // default 1024
	reader        *ReadableStreamDefaultReader
	onStart       Hook
	onPull        Hook
	onCancel      func()
	Controller    *ReadableStreamDefaultController
	CloseChan     chan struct{} // 关闭stream的channel
}

type Hook func(controller *ReadableStreamDefaultController)

var streamIDAcc uint64

func NewReadableStream(options ...ReadableStreamOption) *ReadableStream {
	stream := &ReadableStream{}

	for _, option := range options {
		option(stream)
	}

	if stream.highWaterMark == 0 {
		stream.highWaterMark = 16
	}

	stream.dataChan = make(chan []byte, stream.highWaterMark)
	stream.CloseChan = make(chan struct{})
	stream.reader = &ReadableStreamDefaultReader{stream: stream}
	stream.Controller = &ReadableStreamDefaultController{stream: stream}

	if stream.onStart != nil {
		stream.onStart(stream.Controller)
	}

	atomic.AddUint64(&streamIDAcc, 1)

	stream.ID = fmt.Sprintf("rs-%d", atomic.LoadUint64(&streamIDAcc))
	return stream
}

func (stream *ReadableStream) GetReader() *ReadableStreamDefaultReader {
	return stream.reader
}

func (stream *ReadableStream) Cancel() {

}

func (stream *ReadableStream) Len() int {
	return len(stream.dataChan)
}

type ReadableStreamDefaultController struct {
	stream *ReadableStream
}

func (r *ReadableStreamDefaultController) Enqueue(data []byte) {
	r.stream.dataChan <- data
}

// Close Readers will still be able to read any previously-enqueued chunks from the stream,
// but once those are read, the stream will become closed.
// If you want to completely get rid of the stream and discard any enqueued chunks,
// you'd use ReadableStream.Cancel() or ReadableStreamDefaultReader.Cancel().
func (r *ReadableStreamDefaultController) Close() {
	close(r.stream.CloseChan)
}

type ReadableStreamDefaultReader struct {
	stream *ReadableStream
}

func (reader *ReadableStreamDefaultReader) Read() <-chan []byte {
	return reader.stream.dataChan
}

// Cancel Calling this method signals a loss of interest in the stream by a consumer.
// Cancel is used when you've completely finished with the stream and don't need any more data from it,
// even if there are chunks enqueued waiting to be read.
// That data is lost after cancel is called, and the stream is not readable any more.
// To read those chunks still and not completely get rid of the stream,
// you'd use ReadableStreamDefaultController.Close().
func (reader *ReadableStreamDefaultReader) Cancel() {

}

// Closed https://developer.mozilla.org/en-US/docs/Web/API/ReadableStreamDefaultReader/closed
// The closed read-only property of the ReadableStreamDefaultReader interface returns a Promise that fulfills
// when the stream closes, or rejects if the stream throws an error or the reader's lock is released.
// This property enables you to write code that responds to an end to the streaming process.
func (reader *ReadableStreamDefaultReader) Closed() {

}

func (reader *ReadableStreamDefaultReader) ReleaseLock() {

}

type ReadableStreamOption func(stream *ReadableStream)

func WithHighWaterMark(watermark uint64) ReadableStreamOption {
	return func(stream *ReadableStream) {
		stream.highWaterMark = watermark
	}
}

func WithOnStart(onStart Hook) ReadableStreamOption {
	return func(stream *ReadableStream) {
		stream.onStart = onStart
	}
}

func WithOnPull(onPull Hook) ReadableStreamOption {
	return func(stream *ReadableStream) {
		stream.onPull = onPull
	}
}

func WithOnCancel(onCancel func()) ReadableStreamOption {
	return func(stream *ReadableStream) {
		stream.onCancel = onCancel
	}
}
