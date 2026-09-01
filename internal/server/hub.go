package server

import (
	"fmt"
	"net"
	"tcp-chat/internal/domain"
	"time"
)

type Request[T any] struct {
	Response chan T
}

type Hub struct {
	clients    map[string]*domain.Client
	broadcast  chan domain.ChatMessage
	register   chan *domain.Client
	unregister chan *domain.Client
	requests   chan any
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*domain.Client),
		broadcast:  make(chan domain.ChatMessage),
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

func (h *Hub) Broadcast(m domain.ChatMessage) {
	h.broadcast <- m
}

func (h *Hub) setupClientConnection(conn net.Conn) *domain.Client {
	client := &domain.Client{
		ID:       domain.GenerateClientID(),
		Conn:     conn,
		JoinTime: time.Now(),
	}

	client.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	client.Conn.Write([]byte(fmt.Sprintf("Welcome %s! Type your messages below:\n", client.ID)))

	h.Broadcast(domain.ChatMessage{
		Timestamp:   time.Now(),
		Content:     fmt.Sprintf("Client %s connected from %s", client.ID, conn.RemoteAddr().String()),
		MessageType: domain.MsgTypeSystem,
	})

	return client
}

func (h *Hub) cleanupClient(client *domain.Client) {
	client.Conn.Close()

	h.Unregister(client)

	h.Broadcast(domain.ChatMessage{
		Timestamp:   time.Now(),
		Content:     fmt.Sprintf("Client %s disconnected", client.ID),
		MessageType: domain.MsgTypeSystem,
	})
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

func (h *Hub) broadcastMessage(m domain.ChatMessage) {
	for _, client := range h.clients {
		if client.ID == m.ClientID {
			continue
		}

		client.Conn.Write([]byte(domain.FormatMessage(m) + "\n"))
	}
}
