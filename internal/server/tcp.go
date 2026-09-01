package server

import (
	"github.com/blrrubik/tcp-chat-server/internal/hub"
	"log"
	"net"
)

func StartEchoServer(port string, h *hub.Hub) error {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	h.Logger().Info("server starting", "port", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go h.HandleConnection(conn)
	}
}
