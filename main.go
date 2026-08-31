package main

import (
	"tcp-chat/internal/server"
)

func main() {
	h := server.NewHub()

	go h.Run()

	server.StartEchoServer(":8010", h)
}
