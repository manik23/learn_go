package main

import (
	"context"
	v1 "learn-control-plane/rest/v1"
	"learn-gin/db"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	httpServer := setupServer(port)

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

func setupServer(port string) *http.Server {
	r := gin.Default()

	// Core Endpoints
	setupRoutes(r)

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	return httpServer
}

func setupRoutes(r *gin.Engine) {
	db, err := db.SetupDB()
	if err != nil {
		log.Fatalf("error setting up database: %v", err)
	}
	v1.SetupV1(r, db)
}
