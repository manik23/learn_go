package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	"learn-gin/db"
	v1 "learn-gin/routes/v1"

	"github.com/gin-gonic/gin"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	router := gin.Default()

	db, err := db.SetupDB()
	if err != nil {
		log.Fatal(err)
	}

	if err := v1.SetupV1Routes(router, db); err != nil {
		log.Fatal(err)
	}

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
