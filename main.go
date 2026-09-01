package main

import (
	"tcp-chat/internal/server"
)

func main() {
	logger := server.SetupLogging("info")
	h := server.NewHub(logger)

	go h.Run()

	server.StartEchoServer(":8010", h)
}
