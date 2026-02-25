package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"prod-service-patterns/db"
)

func TestHandleProcess(t *testing.T) {
	ctx := context.Background()
	database, err := db.NewDatabase(ctx, 3) // Small pool for testing
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	appConfig := &AppConfig{
		DB: database,
	}

	handler := newHttpHandler(appConfig)

	t.Run("parallel_requests_exceeding_pool", func(t *testing.T) {
		const numRequests = 10
		var wg sync.WaitGroup
		results := make(chan int, numRequests)

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				req := httptest.NewRequest("GET", "/process", nil)
				rr := httptest.NewRecorder()
				handler.ServeHTTP(rr, req)
				results <- rr.Code
			}()
		}

		wg.Wait()
		close(results)

		var success, timeout int
		for code := range results {
			if code == http.StatusOK {
				success++
			} else if code == http.StatusGatewayTimeout {
				timeout++
			}
		}

		t.Logf("Success: %d, Timeout: %d", success, timeout)
	})
}

func BenchmarkHandleProcess(b *testing.B) {
	ctx := context.Background()
	database, err := db.NewDatabase(ctx, 100)
	if err != nil {
		b.Fatalf("Failed to create database: %v", err)
	}

	appConfig := &AppConfig{
		DB: database,
	}

	handler := newHttpHandler(appConfig)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/process", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
		}
	})
}
