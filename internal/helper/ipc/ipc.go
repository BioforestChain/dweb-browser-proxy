package ipc

import (
	"errors"
	"log"
	"proxyServer/ipc"
	"time"
)

func ReadStreamWithTimeout(stream *ipc.ReadableStream, t time.Duration) ([]byte, error) {
	timer := time.NewTimer(t)

	var timeout bool
	go func() {
		select {
		case <-timer.C:
			timeout = true
			stream.Controller.Close()
		}
	}()

	reader := stream.GetReader()
	data := make([]byte, 0)
	var readErr error
	for {
		r, err := reader.Read()
		if err != nil {
			readErr = err
			log.Println("read err: ", len(data))
			break
		}

		if r.Done {
			log.Println("done: ", len(data))
			break
		}

		data = append(data, r.Value...)
		log.Println("total: ", len(data))
	}

	if timeout {
		return nil, errors.New("read body stream timeout")
	}

	if readErr != nil {
		return nil, readErr
	}

	return data, nil
}
