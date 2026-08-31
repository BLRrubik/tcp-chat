package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func main() {
	StartEchoServer(":8010")
}

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

		fmt.Println("incoming", line)

		conn.Write([]byte(line + "\n"))
	}

	return nil
}
