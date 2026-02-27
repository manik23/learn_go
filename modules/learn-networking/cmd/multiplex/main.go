package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

func main() {
	// 1. Setup a server that can handle many connections
	ln, err := net.Listen("tcp", "localhost:8888")
	if err != nil {
		log.Fatalf("Listen error: %v", err)
	}
	defer ln.Close()

	var activeConns sync.WaitGroup
	connCount := 100 // Let's simulate 100 concurrent "slow" connections

	log.Printf("Starting Multiplexing Demo: 1 Server, %d Clients", connCount)

	// Server: Accept loop
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				// Read once then wait - simulates a slow client
				buf := make([]byte, 1024)
				_, _ = c.Read(buf)
				time.Sleep(2 * time.Second)
				_, _ = c.Write([]byte("Acknowledged\n"))
			}(conn)
		}
	}()

	// Clients: Launch many goroutines
	startTime := time.Now()
	for i := 0; i < connCount; i++ {
		activeConns.Add(1)
		go func(id int) {
			defer activeConns.Done()
			conn, err := net.Dial("tcp", "localhost:8888")
			if err != nil {
				return
			}
			defer conn.Close()

			// Send some data
			fmt.Fprintf(conn, "Hello from client %d\n", id)

			// Wait for reply
			buf := make([]byte, 1024)
			_, _ = conn.Read(buf)
		}(i)
	}

	activeConns.Wait()
	fmt.Printf("Finished handling %d connections in %v\n", connCount, time.Since(startTime))
	fmt.Println("Check 'lsof -nP -i :8888' while this is running to see 2x100 FDs!")
}
