package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"tcp-chat/internal/domain"
	"tcp-chat/internal/storage"
	"time"
)

const shutdownGracePeriod = 10 * time.Second

var logLevels = map[string]int{"error": 0, "warn": 1, "info": 2}

type levelFilterWriter struct {
	dest      io.Writer
	threshold int
}

func (w *levelFilterWriter) Write(p []byte) (int, error) {
	level := logLevels["info"]

	switch {
	case bytes.Contains(p, []byte("ERROR")):
		level = logLevels["error"]
	case bytes.Contains(p, []byte("WARN")):
		level = logLevels["warn"]
	}

	if level > w.threshold {
		return len(p), nil
	}

	return w.dest.Write(p)
}

func SetupLogging(level string) *log.Logger {
	threshold, ok := logLevels[strings.ToLower(level)]
	if !ok {
		threshold = logLevels["info"]
	}

	return log.New(&levelFilterWriter{dest: os.Stdout, threshold: threshold}, "[TCP-CHAT] ", log.Ldate|log.Ltime)
}

type activeClientsRequest struct {
	Response chan []string
}

type clientCountRequest struct {
	Response chan int
}

type messageHistoryRequest struct {
	Response chan []domain.ChatMessage
}

type closeAllRequest struct {
	Response chan struct{}
}

type statsRequest struct {
	Response chan ServerStats
}

type ServerStats struct {
	ActiveConnections      int
	TotalMessagesProcessed int64
	UptimeSeconds          int64
	ErrorCount             int
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
	errors     chan error

	messageHistory *storage.MessageHistory
	logger         *log.Logger
	startedAt      time.Time

	totalMessagesProcessed int64
	errorCount             int

	wg           sync.WaitGroup
	shuttingDown atomic.Bool

	maxConnections int
}

func NewHub(logger *log.Logger, messageHistorySize int, maxConnections int) *Hub {
	return &Hub{
		clients:        make(map[string]*domain.Client),
		broadcast:      make(chan domain.ChatMessage),
		register:       make(chan *domain.Client),
		unregister:     make(chan *domain.Client),
		requests:       make(chan any),
		errors:         make(chan error),
		messageHistory: storage.NewMessageHistory(messageHistorySize),
		logger:         logger,
		startedAt:      time.Now(),
		maxConnections: maxConnections,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.ID] = client
			h.logger.Printf("INFO Client %s connected", client.ID)
		case client := <-h.unregister:
			delete(h.clients, client.ID)
			h.logger.Printf("INFO Client %s disconnected", client.ID)
		case msg, ok := <-h.broadcast:
			if !ok {
				continue
			}

			if msg.MessageType == domain.MsgTypeUser {
				h.messageHistory.Add(msg)
				h.totalMessagesProcessed++
			}

			h.broadcastMessage(msg)
		case err := <-h.errors:
			h.errorCount++
			h.logger.Printf("ERROR %v", err)
		case r := <-h.requests:
			switch req := r.(type) {
			case activeClientsRequest:
				req.Response <- h.activeClients()
			case clientCountRequest:
				req.Response <- len(h.clients)
			case messageHistoryRequest:
				req.Response <- h.messageHistory.GetRecent()
			case closeAllRequest:
				for _, client := range h.clients {
					client.Conn.Close()
				}
				close(req.Response)
			case statsRequest:
				req.Response <- ServerStats{
					ActiveConnections:      len(h.clients),
					TotalMessagesProcessed: h.totalMessagesProcessed,
					UptimeSeconds:          int64(time.Since(h.startedAt).Seconds()),
					ErrorCount:             h.errorCount,
				}
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
	if h.shuttingDown.Load() {
		return
	}

	h.broadcast <- m
}

func (h *Hub) ReportError(err error) {
	h.errors <- err
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

func (h *Hub) IsFull() bool {
	if h.maxConnections <= 0 {
		return false
	}

	return h.GetClientCount() >= h.maxConnections
}

func (h *Hub) GetStats() ServerStats {
	req := statsRequest{Response: make(chan ServerStats)}
	h.requests <- req

	return <-req.Response
}

func (h *Hub) GetMessageHistory() []domain.ChatMessage {
	req := messageHistoryRequest{Response: make(chan []domain.ChatMessage)}
	h.requests <- req

	return <-req.Response
}

func (h *Hub) Shutdown(ctx context.Context) error {
	h.logger.Println("INFO Notifying clients...")
	h.Broadcast(domain.ChatMessage{
		Timestamp:   time.Now(),
		Content:     fmt.Sprintf("Server shutting down in %.0f seconds", shutdownGracePeriod.Seconds()),
		MessageType: domain.MsgTypeSystem,
	})

	h.shuttingDown.Store(true)
	close(h.broadcast)

	h.closeAllConnections()

	done := make(chan struct{})
	go func() {
		h.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		h.logger.Println("INFO All clients disconnected")
	case <-ctx.Done():
		h.logger.Println("WARN Shutdown timeout reached, some goroutines did not finish")
	}

	h.logger.Println("INFO Server stopped gracefully")

	return nil
}

func (h *Hub) closeAllConnections() {
	req := closeAllRequest{Response: make(chan struct{})}
	h.requests <- req
	<-req.Response
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
