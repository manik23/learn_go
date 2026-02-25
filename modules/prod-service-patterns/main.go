package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func newHttpHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/process", http.HandlerFunc(handleProcess))
	return mux
}

func main() {
	// 1. Setup Signal handling for Graceful Shutdown
	// 2. Initialize the http.Server
	// 3. Start server in a goroutine
	// 4. Block here until signal received
	// 5. Trigger server.Shutdown()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	handler := newHttpHandler()
	newHttpServer := &http.Server{
		Addr:    "localhost:8080",
		Handler: handler,
	}

	go func() {
		if err := newHttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", "localhost:8080", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := newHttpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v\n", err)
	}

	log.Println("Server exited properly")

}

func handleProcess(w http.ResponseWriter, r *http.Request) {
	// 1. Create a derived context with a 2-second timeout
	// 2. Call the steps in order: stepAuth -> stepValidate -> stepStore
	// 3. If any step returns an error (including context timeout), return an appropriate HTTP error

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := stepAuth(ctx); err != nil {
		if ctx.Err() != nil {
			http.Error(w, ctx.Err().Error(), http.StatusGatewayTimeout)
			return
		}

		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := stepValidate(ctx); err != nil {
		if ctx.Err() != nil {
			http.Error(w, ctx.Err().Error(), http.StatusGatewayTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := stepStore(ctx); err != nil {
		if ctx.Err() != nil {
			http.Error(w, ctx.Err().Error(), http.StatusGatewayTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Processed successfully"))

}

func stepAuth(ctx context.Context) error {
	// MUST respect ctx.Done()
	select {
	case <-ctx.Done():
		{
			return ctx.Err()
		}
	case <-time.After(time.Duration(rand.Intn(500)) * time.Millisecond):
		{
			// Simulate work
			return nil
		}
	}

}

func stepValidate(ctx context.Context) error {

	// MUST respect ctx.Done()
	select {
	case <-ctx.Done():
		{
			return ctx.Err()
		}
	case <-time.After(time.Duration(rand.Intn(50)) * time.Millisecond):
		{
			// Simulate work
			return nil
		}
	}

}

func stepStore(ctx context.Context) error {

	// MUST respect ctx.Done()

	select {
	case <-ctx.Done():
		{
			return ctx.Err()
		}
	case <-time.After(time.Duration(rand.Intn(2000)) * time.Millisecond):
		{
			// Simulate work
			return nil
		}
	}
}
