package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync/atomic"
	"tcp-chat/internal/message"
	"time"
)

var counter atomic.Uint32

type Client struct {
	ID       string
	conn     net.Conn
	JoinTime time.Time
}

func StartEchoServer(port string) error {
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
		defer conn.Close()

		go HandleClient(&Client{
			ID:       GenerateClientID(),
			conn:     conn,
			JoinTime: time.Now(),
		})
	}

	return nil
}

func HandleClient(client *Client) error {
	scanner := bufio.NewScanner(client.conn)
	for scanner.Scan() {
		line := scanner.Text()

		msg := message.ParseIncomingMessage(line, client.ID)

		client.conn.Write([]byte(message.FormatMessage(msg) + "\n"))
	}

	return nil
}

func GenerateClientID() string {
	counter.Add(1)

	return fmt.Sprintf("User_%d", counter.Load())
}
