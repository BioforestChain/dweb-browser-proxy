package ipc

import (
	"bytes"
	"testing"
)

func TestReadableStream(t *testing.T) {
	stream := NewReadableStream()
	wanted := "hello"
	stream.Controller.Enqueue([]byte(wanted))

	ch := stream.GetReader().Read()
	got := <-ch

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
		stream := NewReadableStream(WithOnStart(func(controller *ReadableStreamDefaultController) {
			controller.Enqueue([]byte(wanted))
		}))

		ch := stream.GetReader().Read()
		got := <-ch

		if string(got) != wanted {
			t.Fatal("onstart failed")
		}
	})

	t.Run("with high water mark", func(t *testing.T) {
		var wanted = "hi"
		stream := NewReadableStream(WithHighWaterMark(2))

		stream.Controller.Enqueue([]byte("hi"))
		stream.Controller.Enqueue([]byte("hi"))

		go func() {
			ch := stream.GetReader().Read()
			got := <-ch

			if string(got) != wanted {
				t.Fatal("onstart failed")
			}
		}()
	})
}

func TestReadableStreamDefaultController_Enqueue(t *testing.T) {

}

func TestReadableStreamDefaultReader_Read(t *testing.T) {
	stream := NewReadableStream()
	data := []byte("abc")
	stream.Controller.Enqueue(data)

	got := make([]byte, 4)
	ch := stream.GetReader().Read()
	got = <-ch

	if !bytes.Equal(got[:3], data) {
		t.Fatal("readablestream reader failed")
	}

}
