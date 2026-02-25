package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"prod-service-patterns/db"
)

// URL of the manually running server
const testURL = "http://localhost:8080/process"

func TestHandleProcessRemote(t *testing.T) {
	// Ensure server is running: go run main.go
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	t.Run("parallel_requests_remote", func(t *testing.T) {
		const numRequests = 20
		var wg sync.WaitGroup
		results := make(chan int, numRequests)

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				resp, err := client.Get(testURL)
				if err != nil {
					t.Errorf("Request failed: %v", err)
					return
				}
				defer resp.Body.Close()
				io.Copy(io.Discard, resp.Body)
				results <- resp.StatusCode
			}()
		}

		wg.Wait()
		close(results)

		var success, timeout, other int
		for code := range results {
			switch code {
			case http.StatusOK:
				success++
			case http.StatusGatewayTimeout:
				timeout++
			default:
				other++
			}
		}

		t.Logf("Remote Results -> Success: %d, Timeout/Busy: %d, Other: %d", success, timeout, other)
	})
}

// BenchmarkHandleProcessRemote measures the overhead of the network stack + server logic.
func BenchmarkHandleProcessRemote(b *testing.B) {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
		},
		Timeout: 5 * time.Second,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(testURL)
			if err != nil {
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkHandleProcessLocal measures ONLY the server logic by calling the handler directly.
// No network stack, no syscalls, just pure Go logic.
func BenchmarkHandleProcessLocal(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup a standalone database pool for the benchmark
	database, _ := db.NewDatabase(ctx, 100)
	appConfig := &AppConfig{
		DB:  database,
		ctx: ctx,
	}

	handler := http.HandlerFunc(appConfig.handleProcess)
	req := httptest.NewRequest("GET", "/process", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(w, req)
		w.Body.Reset()
	}
}
