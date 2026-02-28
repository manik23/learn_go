package main

import (
	"context"
	"learn-gin/db"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "learn-control-plane/rest/v1"

	"github.com/gin-gonic/gin"
)

func main() {
	// Set Gin to release mode if not in debug
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	httpServer := setupServer(ctx, port)

	go func() {
		log.Printf("Control Plane Node active on :%s\n", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error starting server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("error shutting down server: %v", err)
	}
	log.Println("Server exited properly")
}

func setupServer(serverCtx context.Context, port string) *http.Server {
	r := gin.Default()

	// Core Endpoints
	setupRoutes(serverCtx, r)

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	return httpServer
}

func setupRoutes(serverCtx context.Context, r *gin.Engine) {
	db, err := db.SetupDB()
	if err != nil {
		log.Fatalf("error setting up database: %v", err)
	}
	v1.SetupV1(serverCtx, r, db)
}
