package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"tcp-chat/internal/domain"
	"time"
)

const helpText = `Available commands:
  /help  - show this message
  /time  - show current server time
  /users - list active users
  /quit  - disconnect
`

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

		switch line {
		case "/help":
			client.Conn.Write([]byte(helpText))
		case "/time":
			client.Conn.Write([]byte(time.Now().Format(time.RFC1123) + "\n"))
		case "/users":
			h.SendUserList(client)
		case "/quit":
			client.Conn.Write([]byte("Goodbye!\n"))
			h.Unregister(client)

			return
		default:
			h.Broadcast(domain.ParseIncomingMessage(line, client.ID))
		}

		client.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	}

	h.Unregister(client)

	if err := scanner.Err(); err != nil {
		fmt.Println("reading standard input:", err)
	}
}
