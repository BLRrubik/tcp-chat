package server

import (
	"fmt"
	"net"
	"strings"
	"tcp-chat/internal/domain"
	"tcp-chat/internal/storage"
	"time"
)

type activeClientsRequest struct {
	Response chan []string
}

type clientCountRequest struct {
	Response chan int
}

type messageHistoryRequest struct {
	Response chan []domain.ChatMessage
}

const helpText = `Available commands:
  /help  - show this message
  /time  - show current server time
  /users - list active users
  /quit  - disconnect
`

type Hub struct {
	clients    map[string]*domain.Client
	broadcast  chan domain.ChatMessage
	register   chan *domain.Client
	unregister chan *domain.Client
	requests   chan any

	messageHistory *storage.MessageHistory
}

func NewHub() *Hub {
	return &Hub{
		clients:        make(map[string]*domain.Client),
		broadcast:      make(chan domain.ChatMessage),
		register:       make(chan *domain.Client),
		unregister:     make(chan *domain.Client),
		requests:       make(chan any),
		messageHistory: storage.NewMessageHistory(),
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
			if msg.MessageType == domain.MsgTypeUser {
				h.messageHistory.Add(msg)
			}

			h.broadcastMessage(msg)
		case r := <-h.requests:
			switch req := r.(type) {
			case activeClientsRequest:
				req.Response <- h.activeClients()
			case clientCountRequest:
				req.Response <- len(h.clients)
			case messageHistoryRequest:
				req.Response <- h.messageHistory.GetRecent()
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

func (h *Hub) GetActiveClients() []string {
	req := activeClientsRequest{Response: make(chan []string)}
	h.requests <- req

	return <-req.Response
}

func (h *Hub) GetClientCount() int {
	req := clientCountRequest{Response: make(chan int)}
	h.requests <- req

	return <-req.Response
}

func (h *Hub) GetMessageHistory() []domain.ChatMessage {
	req := messageHistoryRequest{Response: make(chan []domain.ChatMessage)}
	h.requests <- req

	return <-req.Response
}

func (h *Hub) setupClientConnection(conn net.Conn) *domain.Client {
	client := &domain.Client{
		ID:       domain.GenerateClientID(),
		Conn:     conn,
		JoinTime: time.Now(),
	}

	client.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	h.SendMessageHistory(client)
	client.Conn.Write([]byte(fmt.Sprintf("Welcome! %d users online. Type /help for commands.\n", h.GetClientCount())))

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

func (h *Hub) SendMessageHistory(client *domain.Client) {
	sb := strings.Builder{}

	sb.WriteString("--- Recent messages ---\n")
	for _, msg := range h.GetMessageHistory() {
		sb.WriteString(domain.FormatMessage(msg) + "\n")
	}
	sb.WriteString("--- End of history ---\n")

	client.Conn.Write([]byte(sb.String()))
}

func (h *Hub) HandleCommand(client *domain.Client, command string) {
	switch command {
	case "/help":
		client.Conn.Write([]byte(helpText))
	case "/time":
		client.Conn.Write([]byte(time.Now().Format(time.RFC1123) + "\n"))
	case "/users":
		h.SendUserList(client)
	case "/quit":
		client.Conn.Write([]byte("Goodbye!\n"))
		client.Conn.Close()
	default:
		h.Broadcast(domain.ParseIncomingMessage(command, client.ID))
	}
}

func (h *Hub) SendUserList(client *domain.Client) {
	sb := strings.Builder{}

	sb.WriteString("Active users:\n")
	for _, id := range h.GetActiveClients() {
		sb.WriteString("\t- " + id + "\n")
	}

	client.Conn.Write([]byte(sb.String()))
}
