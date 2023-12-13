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
	Register chan *Client

	// Unregister requests from the clients.
	Unregister  chan *Client
	EndSyncCond *sync.Cond
	//Shutdown    int32
	//Shutdown chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
		//broadcast:  make(chan []byte),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		EndSyncCond: sync.NewCond(&sync.Mutex{}),
		//Shutdown:    make(chan struct{}),
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
				//plan a
				//h.EndSyncCond.L.Lock()
				////TODO 多个地方读取，广播 sync.Cond 避免使用chan
				//// 1. 发送（广播）信号，让业务方监听到
				//// 2. 业务方根据信号，进行停止
				//fmt.Println("·········································")
				//client.DisConn = true
				//h.EndSyncCond.L.Unlock()
				delete(h.clients, client.ID)
				//h.EndSyncCond.Broadcast()
				//fmt.Println("······················broadcast end~~~~~~")
				//plan b
				//close(h.Shutdown)
				close(client.Shutdown)
				//plan c
				//packed.CancelSrcRelease()
				//packed.InitCtx()

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
