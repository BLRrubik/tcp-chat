package hub

import (
	"tcp-chat/internal/domain"
	"tcp-chat/internal/message"
)

type Request[T any] struct {
	Response chan T
}

type Hub struct {
	clients    map[string]*domain.Client
	broadcast  chan message.ChatMessage
	register   chan *domain.Client
	unregister chan *domain.Client
	requests   chan any
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*domain.Client),
		broadcast:  make(chan message.ChatMessage),
		register:   make(chan *domain.Client),
		unregister: make(chan *domain.Client),
		requests:   make(chan any),
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
			h.broadcastMessage(msg)
		case r := <-h.requests:
			switch req := r.(type) {
			case Request[[]string]:
				req.Response <- h.activeClients()
			case Request[int]:
				req.Response <- len(h.clients)
			}
		}
	}
}

func (h *Hub) Register(c *domain.Client) {
	h.register <- c
}

func (h *Hub) Unregister(c *domain.Client) {
	h.unregister <- c
}

func (h *Hub) Broadcast(m message.ChatMessage) {
	h.broadcast <- m
}

func (h *Hub) GetActiveClients() []string {
	req := Request[[]string]{Response: make(chan []string)}
	h.requests <- req

	return <-req.Response
}

func (h *Hub) GetClientCount() int {
	req := Request[int]{Response: make(chan int)}
	h.requests <- req

	return <-req.Response
}

func (h *Hub) activeClients() []string {
	ids := make([]string, 0, len(h.clients))
	for id := range h.clients {
		ids = append(ids, id)
	}

	return ids
}

func (h *Hub) broadcastMessage(m message.ChatMessage) {
	for _, client := range h.clients {
		if client.ID == m.ClientID {
			continue
		}

		client.Conn.Write([]byte(message.FormatMessage(m) + "\n"))
	}
}
