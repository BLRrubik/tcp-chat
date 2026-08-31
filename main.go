package main

import (
	"tcp-chat/internal/server"
)

func main() {
	hub := server.NewHub()

	go hub.Run()

	server.StartEchoServer(":8010", hub)
}
