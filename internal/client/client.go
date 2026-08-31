package client

import (
	"bufio"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"tcp-chat/internal/domain"
	"tcp-chat/internal/hub"
	"tcp-chat/internal/message"
)

var counter atomic.Uint32

func HandleClient(h *hub.Hub, conn net.Conn, clientID string) {
	client := &domain.Client{
		ID:       clientID,
		Conn:     conn,
		JoinTime: time.Now(),
	}
	defer client.Conn.Close()

	h.Register(client)

	scanner := bufio.NewScanner(client.Conn)
	for scanner.Scan() {
		line := scanner.Text()

		h.Broadcast(message.ParseIncomingMessage(line, client.ID))
	}

	h.Unregister(client)

	if err := scanner.Err(); err != nil {
		fmt.Println("reading standard input:", err)
	}
}

func GenerateClientID() string {
	counter.Add(1)

	return fmt.Sprintf("User_%d", counter.Load())
}
