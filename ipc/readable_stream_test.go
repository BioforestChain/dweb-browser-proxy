package ipc

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestReadableStream(t *testing.T) {
	stream := NewReadableStream()
	wanted := "hello"
	_ = stream.Enqueue([]byte(wanted))

	r, _ := stream.GetReader().Read()
	got := r.Value

	if string(got) != wanted {
		t.Fatal("readablestream implement failed")
	}
}

func TestReadableStream_ID(t *testing.T) {
	stream1 := NewReadableStream()
	stream2 := NewReadableStream()
	if stream1.ID == stream2.ID {
		t.Fatal("readablestream unique ID error")
	}
}

func TestReadableStreamWithOptions(t *testing.T) {
	t.Run("with onstart", func(t *testing.T) {
		var wanted = "hi"
		stream := NewReadableStream(WithOnStart(func(ctrl *ReadableStreamDefaultController) {
			_ = ctrl.Enqueue([]byte(wanted))
			_ = ctrl.Enqueue([]byte(wanted))
		}))

		r, _ := stream.GetReader().Read()
		got := r.Value

		if string(got) != wanted {
			t.Fatal("onstart failed")
		}
	})

	t.Run("with onCancel", func(t *testing.T) {
		var called int
		ch := make(chan struct{})
		stream := NewReadableStream(WithOnCancel(func() {
			called++
			ch <- struct{}{}
		}))
		_ = stream.Cancel()

		<-ch
		if called != 1 {
			t.Fatal("onCancel failed")
		}
	})

	t.Run("with high water mark", func(t *testing.T) {
		var wanted = "hi"
		stream := NewReadableStream(WithHighWaterMark(2))

		_ = stream.Enqueue([]byte("hi"))
		_ = stream.Enqueue([]byte("hi"))

		go func() {
			r, _ := stream.GetReader().Read()
			got := r.Value

			if string(got) != wanted {
				t.Fatal("onstart failed")
			}
		}()
	})
}

func TestReadableStream_onPull(t *testing.T) {
	t.Run("onPull is nil", func(t *testing.T) {
		stream := NewReadableStream(WithHighWaterMark(0))

		var i int
		go func() {
			reader := stream.GetReader()
			for {
				_, _ = reader.Read()
				i++
			}
		}()

		_ = stream.Enqueue([]byte("hi"))
		_ = stream.Enqueue([]byte("hi"))

		time.Sleep(10 * time.Millisecond)

		if i != 2 {
			t.Fatal("onPull failed")
		}
	})

	t.Run("onPull is executed once when initialization", func(t *testing.T) {
		var called int

		_ = NewReadableStream(
			WithHighWaterMark(1),
			WithOnPull(func(ctrl *ReadableStreamDefaultController) {
				called++
			}),
		)

		time.Sleep(time.Millisecond * 1)

		if called != 1 {
			t.Fatal("onPull failed")
		}
	})

	t.Run("onPull is executed when controller enqueue", func(t *testing.T) {
		var called int
		var highWaterMark int = 3

		stream := NewReadableStream(
			WithHighWaterMark(uint64(highWaterMark)),
			WithOnPull(func(ctrl *ReadableStreamDefaultController) {
				called++
			}),
		)
		time.Sleep(time.Millisecond * 10)

		for i := 0; i < highWaterMark-1; i++ {
			_ = stream.Controller.Enqueue([]byte("hi"))
			time.Sleep(time.Millisecond * 10)
		}

		if !(called > 0 && called <= 3) {
			t.Fatal("onPull failed")
		}
	})

	t.Run("onPull is executed when reader read", func(t *testing.T) {
		var called int
		var highWaterMark int = 3

		stream := NewReadableStream(
			WithHighWaterMark(uint64(highWaterMark)),
			WithOnPull(func(ctrl *ReadableStreamDefaultController) {
				called++
			}),
		)

		//time.Sleep(time.Millisecond * 10)
		reader := stream.GetReader()
		for i := 0; i < 2; i++ {
			go func() {
				_, _ = reader.Read()
			}()
		}

		time.Sleep(time.Millisecond * 50)
		if !(called > 0 && called <= 3) {
			t.Fatal("onPull failed")
		}
	})

	t.Run("onPull will be stopped when stream or reader cancel", func(t *testing.T) {
		var called int
		var highWaterMark int = 3

		stream := NewReadableStream(
			WithHighWaterMark(uint64(highWaterMark)),
			WithOnStart(func(ctrl *ReadableStreamDefaultController) {
				for {
					_ = ctrl.Enqueue([]byte("hi"))
				}
			}),
			WithOnPull(func(ctrl *ReadableStreamDefaultController) {
				called++
			}),
		)

		reader := stream.GetReader()
		go func() {
			for {
				_, _ = reader.Read()
			}
		}()

		time.Sleep(time.Millisecond * 10)
		reader.Cancel()
		if !stream.canceled {
			t.Fatal("onPull failed")
		}
	})

	t.Run("onPull will be stopped when controller close", func(t *testing.T) {
		var called int
		var highWaterMark int = 3

		stream := NewReadableStream(
			WithHighWaterMark(uint64(highWaterMark)),
			WithOnStart(func(ctrl *ReadableStreamDefaultController) {
				for {
					_ = ctrl.Enqueue([]byte("hi"))
				}
			}),
			WithOnPull(func(ctrl *ReadableStreamDefaultController) {
				called++
			}),
		)

		reader := stream.GetReader()
		go func() {
			for {
				_, _ = reader.Read()
			}
		}()

		time.Sleep(time.Millisecond * 10)
		stream.Controller.Close()

		if !stream.pullClosed {
			t.Fatal("onPull failed")
		}
	})

	t.Run("onPull will not be executed by default when highWaterMark is 0", func(t *testing.T) {
		var called int

		done := make(chan struct{})

		// highWaterMark==0时，初始化readableStream时，不会触发onPull
		rs := NewReadableStream(
			WithHighWaterMark(0),
			WithOnPull(func(ctrl *ReadableStreamDefaultController) {
				called++
				done <- struct{}{}
			}),
		)

		time.Sleep(time.Millisecond * 10)

		go func() {
			// highWaterMark==0时，入队不会触发onPull
			_ = rs.Enqueue([]byte("hi"))
		}()

		time.Sleep(time.Millisecond * 10)

		if called != 0 {
			t.Fatal("onPull failed")
		}

		// read时会触发onPull，不管highWaterMark值是多少
		_, _ = rs.GetReader().Read()

		<-done

		if called != 1 {
			fmt.Println("called： ", called)
			t.Fatal("onPull failed")
		}
	})

}

func TestReadableStream_GetReader(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("GetReader failed")
		}
	}()
	stream := NewReadableStream()
	stream.GetReader()
	stream.GetReader()
}

func TestReadableStream_Cancel(t *testing.T) {
	t.Run("has a reader", func(t *testing.T) {
		stream := NewReadableStream()
		reader := stream.GetReader()
		if err := stream.Cancel(); !errors.Is(err, ErrCancelStream) {
			t.Fatal("cancel stream failed")
		}

		reader.ReleaseLock()
		if err := stream.Cancel(); err != nil {
			t.Fatal("cancel stream failed")
		}

		if err := stream.Enqueue([]byte("hi")); !errors.Is(err, ErrEnqueue) {
			t.Fatal("cancel stream failed")
		}
	})

	t.Run("no reader", func(t *testing.T) {
		stream := NewReadableStream()
		if err := stream.Cancel(); err != nil {
			t.Fatal("cancel stream failed")
		}

		if err := stream.Enqueue([]byte("hi")); !errors.Is(err, ErrEnqueue) {
			t.Fatal("cancel stream failed")
		}
	})

}

func TestReadableStreamDefaultReader_Read(t *testing.T) {
	stream := NewReadableStream()
	data := []byte("abc")
	_ = stream.Enqueue(data)

	got := make([]byte, 4)
	r, _ := stream.GetReader().Read()
	got = r.Value

	if !bytes.Equal(got[:3], data) {
		t.Fatal("readablestream reader failed")
	}
}

func TestReadableStreamDefaultReader_ReleaseLock(t *testing.T) {
	stream := NewReadableStream()
	data := []byte("abc")
	_ = stream.Enqueue(data)

	reader := stream.GetReader()
	reader.ReleaseLock()

	v, err := reader.Read()
	if !(v == nil && errors.Is(err, ErrReadStream)) {
		t.Fatal("reader release lock failed")
	}

	newReader := stream.GetReader()
	v, err = newReader.Read()
	if !(err == nil && bytes.Equal(v.Value, data)) {
		t.Fatal("reader release lock failed")
	}
}

func TestReadableStreamDefaultReader_Cancel(t *testing.T) {
	stream := NewReadableStream()
	reader := stream.GetReader()

	var failed bool
	var ch = make(chan struct{})
	go func() {
		r, err := reader.Read()
		if !(r.Done && err == nil) {
			failed = true
		}

		if err = stream.Enqueue([]byte("hi")); !errors.Is(err, ErrEnqueue) {
			failed = true
		}

		defer func() {
			close(ch)
			if err := recover(); err == nil {
				failed = true
			}
		}()
		stream.GetReader()

		if err = stream.Cancel(); !errors.Is(err, ErrCancelStream) {
			failed = true
		}
	}()

	reader.Cancel()

	<-ch
	if failed {
		t.Fatal("reader cancel failed")
	}
}

func TestReadableStreamDefaultController_Enqueue(t *testing.T) {
	t.Run("reader cancel", func(t *testing.T) {
		stream := NewReadableStream()
		stream.GetReader().Cancel()

		err := stream.Enqueue([]byte("hi"))
		if !errors.Is(err, ErrEnqueue) {
			t.Fatal("controller enqueue failed when reader cancel")
		}
	})
}

func TestReadableStreamDefaultController_Close(t *testing.T) {
	stream := NewReadableStream()
	data := []byte("hi")
	_ = stream.Enqueue(data)

	stream.Controller.Close()

	err := stream.Enqueue(data)
	if !errors.Is(err, ErrEnqueue) {
		t.Fatal("controller close failed")
	}

	reader := stream.GetReader()
	r, err := reader.Read()
	if err != nil || r.Done || !bytes.Equal(r.Value, data) {
		t.Fatal("controller close failed")
	}

	r, err = reader.Read()
	if err != nil || !r.Done {
		t.Fatal("controller close failed")
	}
}
