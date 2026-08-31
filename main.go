package main

import (
	"tcp-chat/internal/server"
)

func main() {
	server.StartEchoServer(":8010")
}
