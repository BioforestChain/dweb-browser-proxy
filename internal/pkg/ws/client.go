package ws

import (
	ipc2 "github.com/BioforestChain/dweb-browser-proxy/pkg/ipc"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	//writeWait = 10 * time.Second
	writeWait = 1000 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 600 * time.Second
	//pongWait = 7 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	//pingPeriod = 4 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 8
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	ID string

	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound message.
	//send chan []byte

	ipc ipc2.IPC

	inputStream *ipc2.ReadableStream

	closed bool

	mutex sync.Mutex

	Shutdown chan struct{}
}

func NewClient(ID string, hub *Hub, conn *websocket.Conn, ipc ipc2.IPC) *Client {
	return &Client{
		ID:          ID,
		hub:         hub,
		conn:        conn,
		ipc:         ipc,
		inputStream: ipc2.NewReadableStream(),
		Shutdown:    make(chan struct{}),
	}
}

// ReadPump pumps messages from the websocket connection to the hub.
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from the goroutine.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister <- c
		//_ = c.conn.Close()
		c.Close()

		if err := recover(); err != nil {
			// TODO 日志上报
			log.Println("readPump ws panic: ", err)
		}
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(s string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("readPump err: ", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		if err := c.inputStream.Enqueue(message); err != nil {
			log.Println("inputStream Enqueue err: ", err)
		}

		//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		//c.hub.broadcast <- message
	}
}

// WritePump pumps messages from the hub to the websocket connection.
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	reader := c.ipc.GetOutputStreamReader()

	defer func() {
		ticker.Stop()
		//_ = c.conn.Close()
		c.Close()
		reader.Cancel()

		if err := recover(); err != nil {
			// TODO 日志上报
			log.Println("writePump ws panic: ", err)
		}
	}()

	go func() {
		for {
			<-ticker.C
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				reader.Cancel()
				return
			}
		}
	}()

	for {
		message, err := reader.Read()
		_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err != nil || message.Done {
			// The hub closed the channel.
			_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		w, err := c.conn.NextWriter(websocket.BinaryMessage)
		if err != nil {
			return
		}

		_, _ = w.Write(message.Value)

		// Add queued messages to the current websocket message.
		//n := len(c.send)
		//for i := 0; i < n; i++ {
		//	_, _ = w.Write(newline)
		//	_, _ = w.Write(<-c.send)
		//}

		if err := w.Close(); err != nil {
			log.Println("writePump close: ", err)
			return
		}
	}
}

func (c *Client) GetHub() *Hub {
	return c.hub
}

func (c *Client) GetIpc() ipc2.IPC {
	return c.ipc
}

func (c *Client) GetInputStream() *ipc2.ReadableStream {
	return c.inputStream
}

func (c *Client) Online() bool {
	return c.hub.Online(c.ID)
}

func (c *Client) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return
	}
	c.closed = true

	c.ipc.Close()
	c.inputStream.Controller.Close()
	_ = c.conn.Close()
}
