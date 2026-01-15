package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"
)

const (
	WORKER = 2
)

type Job struct {
	ID   int    `json:"id"`
	Data string `json:"data"`
}

type Processor interface {
	Process(ctx context.Context, j Job) error
}

type simpleProcessor struct{}

func (s *simpleProcessor) Process(ctx context.Context, j Job) error {

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		time.Sleep(time.Microsecond * 1)
		log.Println("Did something :", j.ID, " ", j.Data)
		return nil
	}

}

func newServer(queue chan Job, success *uint64, failure *uint64) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/submitX", func(w http.ResponseWriter, r *http.Request) {
		var j Job
		if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		select {
		case queue <- j:
			w.WriteHeader(http.StatusAccepted)
		default:
			atomic.AddUint64(failure, 1)
			log.Println("queue full")
			http.Error(w, "queue full", http.StatusServiceUnavailable)
		}
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		lenQeue := len(queue)
		_ = json.NewEncoder(w).Encode(map[string]any{"queue_Depth": lenQeue, "http_success": success, "http_failure": failure})
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hello := "hello"
		w.Write([]byte(hello))
		w.WriteHeader(http.StatusAccepted)

	})

	mux.HandleFunc("/:name", func(w http.ResponseWriter, r *http.Request) {

	})

	return &http.Server{
		Handler: mux,
	}
}

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	queue := make(chan Job, 100)
	s := simpleProcessor{}

	var wg sync.WaitGroup

	var success uint64
	var failure uint64

	for range WORKER {
		wg.Go(func() {
			for j := range queue {
				jobCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				if err := s.Process(jobCtx, j); err != nil {
					atomic.AddUint64(&failure, 1)
					log.Print("process:", err.Error())
				}
				cancel()
				atomic.AddUint64(&success, 1)
			}
		})
	}

	srv := newServer(queue, &success, &failure)
	srv.Addr = ":8080"
	go func() {
		log.Println("Starting the HTTP server")
		srv.ListenAndServe()
	}()

	// Wait fotr shutdown
	<-ctx.Done()

	// Stop Accepting new requests
	srv.Shutdown(context.Background())
	close(queue)
	wg.Wait()
	log.Print("ShutDown Complete")

}
