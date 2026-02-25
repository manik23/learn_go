package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/pprof"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"prod-service-patterns/db"
)

const (
	CAPACITY            = 10
	jobIDKey contextKey = "job_id"
)

var (
	jobID          uint64 = 0
	successfulResp        = []byte("Processed successfully")
)

type contextKey string

type AppConfig struct {
	DB  *db.Database
	ctx context.Context
}

func newHttpHandler(appConfig *AppConfig) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/process", http.HandlerFunc(appConfig.handleProcess))
	return mux
}

func setupPProf(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/heap", pprof.Index)
	mux.HandleFunc("/debug/pprof/goroutine", pprof.Index)
	mux.HandleFunc("/debug/pprof/block", pprof.Index)
	mux.HandleFunc("/debug/pprof/threadcreate", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Index)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Index)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	server := &http.Server{Addr: "127.0.0.1:9000", Handler: mux}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", "127.0.0.1:9000", err)
		}
	}()

	<-ctx.Done()

	log.Println("Shutting down pprof server...")

	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server forced to shutdown: %v\n", err)
	}

}

func main() {
	// 1. Setup Signal handling for Graceful Shutdown
	// 2. Initialize the http.Server
	// 3. Start server in a goroutine
	// 4. Block here until signal received
	// 5. Trigger server.Shutdown()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go setupPProf(ctx)

	db, err := db.NewDatabase(ctx, CAPACITY)
	if err != nil {
		log.Fatalf("Failed to setup database: %v\n", err)
	}

	appConfig := &AppConfig{
		DB:  db,
		ctx: ctx,
	}

	handler := newHttpHandler(appConfig)
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

func (appConfig *AppConfig) handleProcess(w http.ResponseWriter, r *http.Request) {
	// 1. Create a derived context with a 2-second timeout
	// 2. Call the steps in order: stepAuth -> stepValidate -> stepStore
	// 3. If any step returns an error (including context timeout), return an appropriate HTTP error

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	ctx = context.WithValue(ctx, jobIDKey, atomic.AddUint64(&jobID, 1))

	if err := appConfig.stepAuth(ctx); err != nil {
		if ctx.Err() != nil {
			http.Error(w, ctx.Err().Error(), http.StatusGatewayTimeout)
			return
		}

		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := appConfig.stepValidate(ctx); err != nil {
		if ctx.Err() != nil {
			http.Error(w, ctx.Err().Error(), http.StatusGatewayTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := appConfig.stepStore(ctx); err != nil {
		if ctx.Err() != nil {
			http.Error(w, ctx.Err().Error(), http.StatusGatewayTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(successfulResp)
}

func (appConfig *AppConfig) stepAuth(ctx context.Context) error {
	t := time.NewTimer(time.Duration(rand.Intn(500)) * time.Millisecond)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (appConfig *AppConfig) stepValidate(ctx context.Context) error {
	t := time.NewTimer(time.Duration(rand.Intn(50)) * time.Millisecond)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (appConfig *AppConfig) stepStore(ctx context.Context) error {
	// MUST respect ctx.Done()

	select {
	case <-ctx.Done():
		{
			// User centric cancellation
			return ctx.Err()
		}
	case <-appConfig.ctx.Done():
		{
			// Application centric cancellation
			return appConfig.ctx.Err()
		}
	case <-appConfig.DB.Token:
		{
			defer func() {
				appConfig.DB.Token <- struct{}{}
			}()

			select {
			case <-ctx.Done(): // User left or Timeout
				return ctx.Err()
			case <-appConfig.ctx.Done(): // Global Shutdown
				return appConfig.ctx.Err()
			default:
				user := db.User{
					Name:  fmt.Sprintf("User-%v", ctx.Value(jobIDKey)),
					Email: fmt.Sprintf("user-%v@example.com", ctx.Value(jobIDKey)),
				}

				if err := appConfig.DB.DB.WithContext(ctx).Create(&user).Error; err != nil {
					return err
				}
				return nil
			}
		}
	}
}
