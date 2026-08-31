package server

import (
	"log"
	"net"
	"sync/atomic"
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

		go handleClient(conn, GenerateClientID())
	}

	return nil
}
