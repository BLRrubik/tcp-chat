package main

import (
	"tcp-chat/internal/hub"
	"tcp-chat/internal/server"
)

func main() {
	h := hub.NewHub()

	go h.Run()

	server.StartEchoServer(":8010", h)
}
