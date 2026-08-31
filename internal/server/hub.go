package server

import "tcp-chat/internal/message"

type Hub struct {
	clients    map[string]*Client
	broadcast  chan message.ChatMessage
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan message.ChatMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.ID] = client
		case client := <-h.unregister:
			delete(h.clients, client.ID)
		case msg := <-h.broadcast:
			h.BroadcastMessage(msg)
		}
	}
}

func (h *Hub) BroadcastMessage(m message.ChatMessage) {
	for _, client := range h.clients {
		if client.ID == m.ClientID {
			continue
		}

		client.Conn.Write([]byte(message.FormatMessage(m) + "\n"))
	}
}
