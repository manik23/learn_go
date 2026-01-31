package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"testing"
	"time"
)

// “How fast can I create 100 goroutines, serialize JSON, and hammer localhost?”
func BenchmarkSubmt(b *testing.B) {
	client := http.Client{Timeout: 5 * time.Second}

	for b.Loop() {
		var wg sync.WaitGroup
		wg.Add(100)
		for c := 0; c < 100; c++ {
			go func() {
				defer wg.Done()
				j := Job{
					ID:   1,
					Data: "test",
				}
				body, _ := json.Marshal(j)

				client.Post("http://localhost:8080/submitX", "application/json", bytes.NewReader(body))
			}()
		}
		wg.Wait()

	}
}

// How fast can my system process one request under bounded concurrency?”
// go test -benchmem -run=^$ -bench ^BenchmarkSubmit$ learn-routines -count=5
// go test -benchmem -run=<avoid other test> -bench ^<regexofTestName>$ <moduleName> -count=<iterations>

func BenchmarkSubmit(b *testing.B) {
	// --- Disable logging (CRITICAL) ---
	log.SetOutput(io.Discard)

	// --- Start real server ---
	time.Sleep(200 * time.Millisecond) // allow server to start

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
		},
	}

	job := Job{ID: 1, Data: "test"}

	// --- Reusable buffer pool ---
	bufPool := sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}

	// --- Fixed concurrency (do NOT exceed workers too much) ---
	const concurrency = 100
	sem := make(chan struct{}, concurrency)

	b.ResetTimer()

	for b.Loop() {
		sem <- struct{}{}

		buf := bufPool.Get().(*bytes.Buffer)
		buf.Reset()

		_ = json.NewEncoder(buf).Encode(job)

		req, _ := http.NewRequest(
			http.MethodPost,
			"http://localhost:8080/submitX",
			buf,
		)
		req.Header.Set("Content-Type", "application/json")

		_, err := client.Do(req)
		if err != nil {
			b.Fatal(err)
		}

		bufPool.Put(buf)
		<-sem
	}
}
