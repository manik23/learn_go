package main

import (
	"context"
	v1 "learn-gin/routes/v1"

	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	router := gin.Default()

	v1.SetupV1Routes(router)

	server := &http.Server{
		Addr:    ":8081",
		Handler: router.Handler(),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Printf("error : listen: %s\n", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server")
	if err := server.Shutdown(ctx); err != nil {
		log.Println("error in shutting down", err.Error())
	}

	log.Println("exiting service")
}
