package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer ln.Close()

	log.Printf("TCP Echo Server listening on port %s", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	// Intentional Leak for Experimentation

	/*

		CLOSE_WAIT (The Leak)
		When you see CLOSE_WAIT, it means the Remote side (client connection) has closed, the local Kernel (server kernel) has acknowledged it with an ACK, and now the Local Kernel is waiting for the Local Application to call Close().

		Because our server (in LEAK=true mode) returns from the function without calling conn.Close(), the File Descriptor stays open in the process table.
		The Kernel cannot send the final FIN to finish the handshake because the application hasn't "given up" the socket yet.

	*/
	if os.Getenv("LEAK") == "true" {
		log.Printf("[LEAK] New connection from %s - NOT closing FD!", conn.RemoteAddr().String())
		return // Function returns, but connection is NEVER closed
	}

	defer conn.Close()
	remoteAddr := conn.RemoteAddr().String()
	log.Printf("New connection from %s", remoteAddr)

	reader := bufio.NewReader(conn)
	for {
		// Read until newline
		message, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from %s: %v", remoteAddr, err)
			}
			break
		}

		fmt.Printf("[%s] Received: %s", remoteAddr, message)

		// Echo back
		_, err = conn.Write([]byte("ECHO: " + message))
		if err != nil {
			log.Printf("Error writing to %s: %v", remoteAddr, err)
			break
		}
	}

	log.Printf("Connection closed: %s", remoteAddr)
}
