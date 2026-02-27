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
	for i := 0; i < 2000; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, err := net.Dial("tcp", serverAddr)
			if err != nil {
				log.Printf("[%d] Dial error: %v", id, err)
				return
			}
			/*
				Close immediately to force TIME_WAIT
				TIME_WAIT (The Zombie)
				This state occurs on the side that initiates the connection teardown (the Active Closer).
				It stays in TIME_WAIT for 2MSL (Maximum Segment Lifetime, usually 60-120s) to ensure:
				1. Reliability: The final ACK is received by the peer.
				2. Safety: Old packets from this connection don't interfere with new connections (Delayed Duplicates).

				In this "Bomb" program, we are intentionally creating thousands of connections and closing them immediately.
				This floods the kernel's connection table with sockets stuck in TIME_WAIT, preventing new connections from being established.

			*/

			/*
				$ LEAK=true make run-server

				$ make run-bomb

				The FIN_WAIT_2 (The Hang) when server leaking connections
				The side that initiated the close (the active closer) is sitting in FIN_WAIT_2.
				It has sent its FIN, received an ACK, and is now waiting for the other side to send its FIN.
				It will stay here until a timeout occurs (often 60 seconds) because the other side (server)is stuck in CLOSE_WAIT.
			*/
			conn.Close()
		}(i)

		// Tiny sleep to avoid overloading local CPU, but fast enough to fill ports
		time.Sleep(2 * time.Millisecond)
	}

	wg.Wait()
	log.Println("Bomb finished. Check 'netstat -an | grep 8080 | grep TIME_WAIT | wc -l'")
}
