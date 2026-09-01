package main

import (
	"log"
	"os"
	"tcp-chat/internal/server"
)

func main() {
	logger := SetupLogging("info")
	h := server.NewHub(logger)

	go h.Run()

	server.StartEchoServer(":8010", h)
}

func SetupLogging(level string) *log.Logger {
	return log.New(os.Stdout, "[TCP-CHAT] ", log.Ldate|log.Ltime)
}
