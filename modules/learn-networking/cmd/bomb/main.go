package main

import (
	"log"
	"net"
	"sync"
	"time"
)

func main() {
	serverAddr := "localhost:8080"
	var wg sync.WaitGroup

	log.Printf("Starting Port Bomb against %s...", serverAddr)

	// Attempt to open/close 2000 connections as fast as possible
	for i := 0; i < 2000000; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, err := net.Dial("tcp", serverAddr)
			if err != nil {
				log.Printf("[%d] Dial error: %v", id, err)
				return
			}
			// Close immediately to force TIME_WAIT
			conn.Close()
		}(i)

		// Tiny sleep to avoid overloading local CPU, but fast enough to fill ports
		time.Sleep(2 * time.Millisecond)
	}

	wg.Wait()
	log.Println("Bomb finished. Check 'netstat -an | grep 8080 | grep TIME_WAIT | wc -l'")
}
