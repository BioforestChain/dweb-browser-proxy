package ws

import (
	"log"
)

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients.
	clients map[string]*Client

	// Inbound messages from the clients.
	//broadcast chan []byte

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from the clients.
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
		//broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.clients[client.ID] = client
			log.Println("ws hub: ", h.clients)
		case client := <-h.Unregister:
			if _, ok := h.clients[client.ID]; ok {
				//结束,发送信号
				delete(h.clients, client.ID)
				close(client.Shutdown)
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
