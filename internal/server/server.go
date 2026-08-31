package server

import (
	"log"
	"net"

	"tcp-chat/internal/client"
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

		go client.HandleClient(h, conn, client.GenerateClientID())
	}
}
