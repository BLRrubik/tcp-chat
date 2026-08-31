package server

import (
	"bufio"
	"fmt"
	"net"
	"tcp-chat/internal/message"
	"time"
)

func handleClient(conn net.Conn, clientID string) {
	client := &Client{
		ID:       clientID,
		Conn:     conn,
		JoinTime: time.Now(),
	}
	defer client.Conn.Close()

	scanner := bufio.NewScanner(client.Conn)
	for scanner.Scan() {
		line := scanner.Text()

		msg := message.ParseIncomingMessage(line, client.ID)

		client.Conn.Write([]byte(message.FormatMessage(msg) + "\n"))
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("reading standard input:", err)
	}
}

func GenerateClientID() string {
	counter.Add(1)

	return fmt.Sprintf("User_%d", counter.Load())
}
