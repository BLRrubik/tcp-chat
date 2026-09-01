package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"tcp-chat/internal/domain"
	"time"
)

func StartEchoServer(port string, h *Hub) error {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	h.logger.Printf("INFO Server starting on port %s", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleClient(h, conn)
	}
}

func handleClient(h *Hub, conn net.Conn) {
	var client *domain.Client

	defer func() {
		if r := recover(); r != nil {
			id := "unknown"
			if client != nil {
				id = client.ID
			}

			h.logger.Printf("WARN Cleaning up %s, server continues", id)
			h.ReportError(fmt.Errorf("panic in client %s: %v", id, r))
		}

		if client != nil {
			h.cleanupClient(client)
		}
	}()

	client = h.setupClientConnection(conn)
	h.Register(client)

	scanner := bufio.NewScanner(client.Conn)
	for scanner.Scan() {
		line := scanner.Text()

		h.HandleCommand(client, line)

		client.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	}

	h.Unregister(client)

	if err := scanner.Err(); err != nil {
		h.logger.Printf("ERROR Client %s connection error: %v", client.ID, err)
		h.ReportError(err)
	}
}
