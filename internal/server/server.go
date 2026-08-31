package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"tcp-chat/internal/message"
)

func StartEchoServer(port string, h *Hub) error {
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

		go handleClient(h, conn)
	}
}

func handleClient(h *Hub, conn net.Conn) {
	client := h.setupClientConnection(conn)
	defer h.cleanupClient(client)

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
