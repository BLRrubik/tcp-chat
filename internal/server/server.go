package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"tcp-chat/internal/domain"
	"tcp-chat/internal/message"
	"time"

	"tcp-chat/internal/hub"
)

func StartEchoServer(port string, h *hub.Hub) error {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleClient(h, conn, domain.GenerateClientID())
	}
}

func handleClient(h *hub.Hub, conn net.Conn, clientID string) {
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
