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
	Conn     net.Conn
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
			Conn:     conn,
			JoinTime: time.Now(),
		})
	}

	return nil
}

func HandleClient(client *Client) error {
	scanner := bufio.NewScanner(client.Conn)
	for scanner.Scan() {
		line := scanner.Text()

		msg := message.ParseIncomingMessage(line, client.ID)

		client.Conn.Write([]byte(message.FormatMessage(msg) + "\n"))
	}

	return nil
}

func GenerateClientID() string {
	counter.Add(1)

	return fmt.Sprintf("User_%d", counter.Load())
}
