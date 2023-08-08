package ws

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients.
	clients map[string]*Client

	// Inbound messages from the clients.
	//broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from the clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
		//broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.ID] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.send)
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
