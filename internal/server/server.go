package server

import (
	"bufio"
	"log"
	"net"
	"tcp-chat/internal/message"
)

func StartEchoServer(port string) error {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	conn, err := listener.Accept()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()

		msg := message.ParseIncomingMessage(line, conn.RemoteAddr().String())

		conn.Write([]byte(message.FormatMessage(msg) + "\n"))
	}

	return nil
}
