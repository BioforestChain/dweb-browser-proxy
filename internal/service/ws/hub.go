package service

import (
	"log"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients.
	clients map[string]*Client

	// Inbound messages from the clients.
	//broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from the clients.
	unregister  chan *Client
	EndSyncCond *sync.Cond
	//Shutdown    int32
	Shutdown chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
		//broadcast:  make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		EndSyncCond: sync.NewCond(&sync.Mutex{}),
		Shutdown:    make(chan struct{}),
	}
}

func (h *Hub) Run() {

	for {
		select {
		case client := <-h.register:
			h.clients[client.ID] = client
			log.Println("ws hub: ", h.clients)
		case client := <-h.unregister:
			if _, ok := h.clients[client.ID]; ok {
				//结束,发送信号
				//h.EndSyncCond.L.Lock()
				////atomic.StoreInt32(&h.Shutdown, 1)
				//h.Shutdown <- struct{}{}
				//h.EndSyncCond.Signal()
				//h.EndSyncCond.L.Unlock()
				delete(h.clients, client.ID)
				//close(client.send)
			}
			//case message := <-h.broadcast:
			//	for _, client := range h.clients {
			//		select {
			//		case client.send <- message:
			//		default:
			//			close(client.send)
			//			delete(h.clients, client.ID)
			//		}
			//	}
		}
	}
}

func (h *Hub) GetClient(clientID string) *Client {
	client, ok := h.clients[clientID]
	if !ok {
		return nil
	}
	return client
}
