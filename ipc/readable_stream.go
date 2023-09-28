package ipc

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
)

var ErrClosedStream = errors.New("this readable stream has been closed")
var ErrEnqueue = errors.New("failed to execute 'enqueue' on 'ReadableStreamDefaultController': " +
	"Cannot enqueue a chunk into a closed readable stream")
var ErrReadStream = errors.New("failed to execute 'read' on 'ReadableStreamDefaultReader': " +
	"This readable stream reader has been released and cannot be used to read from its previous owner stream")
var ErrCancelStream = errors.New("failed to execute 'cancel' on 'ReadableStream': Cannot cancel a locked stream")
var ErrGetReader = errors.New("failed to execute 'getReader' on 'ReadableStream': ReadableStreamDefaultReader constructor " +
	"can only accept readable streams that are not yet locked to a reader")

// ReadableStream Streams API的不完全实现 https://developer.mozilla.org/en-US/docs/Web/API/Streams_API
// backpressure/queuing strategies机制未实现
type ReadableStream struct {
	ID            string
	dataChan      chan []byte
	closed        bool    // true when close(dataChan)
	highWaterMark *uint64 // default 10
	readerLocked  bool
	mu            sync.Mutex
	onStart       Hook // 使用onStart要自行实现退出机制，否则可能会出现goroutine泄露，
	onPull        Hook // controller.enqueue(xx) 或 reader.read()时，都会触发执行
	onCancel      func()
	Controller    *ReadableStreamDefaultController
	cancelChan    chan struct{} // used when stream or reader calls Cancel
	canceled      bool          // true when stream or reader Cancel
	pullChan      chan int      // 0表示enqueue时，触发pull；1表示read时，触发pull
	pullClosed    bool
}

type Hook func(controller *ReadableStreamDefaultController)

var streamIDAcc uint64
var defaultHighWaterMark uint64 = 10

func NewReadableStream(options ...ReadableStreamOption) *ReadableStream {
	stream := &ReadableStream{}

	for _, option := range options {
		option(stream)
	}

	if stream.highWaterMark == nil {
		stream.highWaterMark = &defaultHighWaterMark
	}

	stream.dataChan = make(chan []byte, *stream.highWaterMark)
	stream.cancelChan = make(chan struct{})
	stream.Controller = &ReadableStreamDefaultController{stream: stream}

	if stream.onStart != nil {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Println("ReadableStream start panic: ", err)
				}
			}()

			stream.onStart(stream.Controller)
		}()
	}

	if stream.onPull != nil {
		stream.pullChan = make(chan int, 1)

		go func() {
			stream.pulling(0)

			for {
				select {
				case <-stream.cancelChan:
					return
				case trigger, ok := <-stream.pullChan:
					if !ok {
						return
					}
					stream.pulling(trigger)
				}
			}
		}()
	}

	if stream.onCancel != nil {
		go func() {
			select {
			case <-stream.cancelChan:
				stream.onCancel()
			}
		}()
	}

	atomic.AddUint64(&streamIDAcc, 1)

	stream.ID = fmt.Sprintf("rs-%d", atomic.LoadUint64(&streamIDAcc))
	return stream
}

func (stream *ReadableStream) GetReader() ReadableStreamReader {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	if stream.readerLocked {
		panic("Failed to execute 'GetReader' on 'ReadableStream': ReadableStreamDefaultReader constructor " +
			"can only accept readable streams that are not yet locked to a reader")
	}

	stream.readerLocked = true
	return &ReadableStreamDefaultReader{stream: stream}
}

// Cancel is used when you've completely finished with the stream and don't need any more data from it,
// even if there are chunks enqueued waiting to be read.
// That data is lost after cancel is called, and the stream is not readable any more.
// To read those chunks still and not completely get rid of the stream,
// you'd use ReadableStreamDefaultController.Close().
func (stream *ReadableStream) Cancel() error {
	if stream.readerLocked {
		return ErrCancelStream
	}

	stream.close()
	stream.cancel()
	return nil
}

func (stream *ReadableStream) Len() int {
	return len(stream.dataChan)
}

func (stream *ReadableStream) Enqueue(data []byte) (err error) {
	return stream.Controller.Enqueue(data)
}

func (stream *ReadableStream) close() {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	if !stream.closed {
		stream.closed = true
		close(stream.dataChan)
	}
}

func (stream *ReadableStream) cancel() {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	if !stream.canceled {
		stream.canceled = true
		close(stream.cancelChan)
	}
}

func (stream *ReadableStream) pull(trigger int) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	if stream.pullChan != nil && len(stream.pullChan) == 0 && !stream.pullClosed {
		stream.pullChan <- trigger
	}
}

func (stream *ReadableStream) pulling(trigger int) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("ReadableStream pull panic: ", err)
		}
	}()

	v := stream.Controller.desiredSize()
	if trigger == 1 || (v > 0 && stream.onPull != nil) {
		stream.onPull(stream.Controller)
	}
}

func (stream *ReadableStream) stopPull() {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	if stream.pullChan != nil && !stream.pullClosed {
		close(stream.pullChan)
		stream.pullClosed = true
	}
}

type ReadableStreamDefaultController struct {
	stream *ReadableStream
}

func (ctrl *ReadableStreamDefaultController) Enqueue(data []byte) (err error) {
	defer func() {
		if er := recover(); er != nil {
			err = ErrEnqueue
		}
		return
	}()

	ctrl.stream.dataChan <- data

	ctrl.stream.pull(0)
	return
}

// Close Readers will still be able to read any previously-enqueued chunks from the stream,
// but once those are read, the stream will become closed.
// If you want to completely get rid of the stream and discard any enqueued chunks,
// you'd use ReadableStream.Cancel() or ReadableStreamDefaultReader.Cancel().
func (ctrl *ReadableStreamDefaultController) Close() {
	ctrl.stream.close()
	ctrl.stream.stopPull()
}

func (ctrl *ReadableStreamDefaultController) desiredSize() int {
	return int(*ctrl.stream.highWaterMark) - ctrl.stream.Len()
}

type ReadableStreamReader interface {
	// Read the next chunk in the stream's internal queue.
	Read() (*StreamResult, error)

	// Cancel Calling this method signals a loss of interest in the stream by a consumer.
	// Cancel is used when you've completely finished with the stream and don't need any more data from it,
	// even if there are chunks enqueued waiting to be read.
	// That data is lost after cancel is called, and the stream is not readable any more.
	// To read those chunks still and not completely get rid of the stream,
	// you'd use ReadableStreamDefaultController.Close().
	//
	// If the reader is active, the cancel() method behaves the same
	// as that for the associated stream (ReadableStream.Cancel()).
	Cancel()

	ReleaseLock()
}

type ReadableStreamDefaultReader struct {
	stream   *ReadableStream
	released bool // whether to release a stream
}

type StreamResult struct {
	Done  bool
	Value []byte
}

func (reader *ReadableStreamDefaultReader) Read() (*StreamResult, error) {
	if reader.released {
		return nil, ErrReadStream
	}

	if reader.stream.canceled {
		return &StreamResult{Done: true, Value: nil}, nil
	}

	reader.stream.pull(1)

	data, ok := <-reader.stream.dataChan
	if !ok {
		reader.stream.stopPull()
		return &StreamResult{Done: true, Value: nil}, nil
	}

	return &StreamResult{Done: false, Value: data}, nil
}

func (reader *ReadableStreamDefaultReader) Cancel() {
	reader.stream.close()
	reader.stream.cancel()
}

// Closed https://developer.mozilla.org/en-US/docs/Web/API/ReadableStreamDefaultReader/closed
// The closed read-only property of the ReadableStreamDefaultReader interface returns a Promise that fulfills
// when the stream closes, or rejects if the stream throws an error or the reader's lock is released.
// This property enables you to write code that responds to an end to the streaming process.
func (reader *ReadableStreamDefaultReader) Closed() {

}

func (reader *ReadableStreamDefaultReader) ReleaseLock() {
	reader.stream.mu.Lock()
	defer reader.stream.mu.Unlock()

	reader.stream.readerLocked = false
	reader.released = true
}

type ReadableStreamOption func(stream *ReadableStream)

func WithHighWaterMark(watermark uint64) ReadableStreamOption {
	return func(stream *ReadableStream) {
		stream.highWaterMark = &watermark
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
