package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	serverAddr := "localhost:8080"
	if addr := os.Getenv("SERVER_ADDR"); addr != "" {
		serverAddr = addr
	}

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	log.Printf("Connected to %s", serverAddr)

	reader := bufio.NewReader(os.Stdin)
	serverReader := bufio.NewReader(conn)

	for {
		fmt.Print("Text to send: ")
		text, _ := reader.ReadString('\n')

		// Exit on empty line or "exit"
		if text == "\n" || text == "exit\n" {
			conn.Close()
			break
		}

		fmt.Fprintf(conn, text)

		message, err := serverReader.ReadString('\n')
		if err != nil {
			log.Fatalf("Server closed connection: %v", err)
		}
		fmt.Printf("Server: %s", message)
	}
}
