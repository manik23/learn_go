package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	WORKER_FACTOR = 2
	RETRIES       = 3
)

var (
	stats = make(map[int]int)
	mu    = sync.RWMutex{}
)

type Job struct {
	ID   int    `json:"id"`
	Data string `json:"data"`
}

type Result struct {
	JobID int
	Len   int
}

type Processor interface {
	Process(ctx context.Context, j Job, t *time.Timer, workTime time.Duration) error
}

type simpleProcessor struct{}

func (s *simpleProcessor) Process(ctx context.Context, j Job, t *time.Timer, workTime time.Duration) error {

	t.Reset(workTime)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		if rand.Intn(10) < 3 {
			return errors.New("random database error")
		}
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

		mu.RLock()
		statsCopy := make(map[int]int)
		for k, v := range stats {
			statsCopy[k] = v
		}
		lenQeue := len(queue)
		_ = json.NewEncoder(w).Encode(map[string]any{"queue_Depth": lenQeue, "http_success": success, "http_failure": failure, "jobs_done": statsCopy})
		mu.RUnlock()
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hello := "hello"
		w.Write([]byte(hello))
		w.WriteHeader(http.StatusAccepted)
	})

	return &http.Server{
		Handler: mux,
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	queue := make(chan Job, 1000)
	results := make(chan Result, 1000)

	s := simpleProcessor{}

	var wg sync.WaitGroup

	var success uint64
	var failure uint64

	for range WORKER_FACTOR * runtime.NumCPU() {

		wg.Add(1)
		go func() {
			defer wg.Done()
			// PRE-ALLOCATED: One timer per worker
			// Initialized with a long duration; it will be Reset later
			t := time.NewTimer(time.Hour)
			defer t.Stop()

			for j := range queue {
				jobCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
				var err error
				for retry := range RETRIES {
					workTime := time.Duration(rand.Intn(20)) * time.Millisecond
					if err = s.Process(jobCtx, j, t, workTime); err == nil {
						break
					}
					time.Sleep((1 << retry) * time.Millisecond)
				}
				if err != nil {
					log.Print("process failed after retries:", err.Error())
					atomic.AddUint64(&failure, 1)
				} else {
					atomic.AddUint64(&success, 1)
					results <- Result{JobID: j.ID, Len: len(j.Data)}
				}
				cancel()
			}
		}()
	}

	srv := newServer(queue, &success, &failure)
	srv.Addr = ":8080"
	go func() {
		log.Println("Starting the HTTP server at 8080")
		srv.ListenAndServe()
	}()

	var aggWG sync.WaitGroup
	aggWG.Add(1)
	go func() {
		defer aggWG.Done()
		for r := range results {
			mu.Lock()
			stats[r.JobID] += r.Len
			mu.Unlock()
		}
	}()

	// Wait for shutdown
	<-ctx.Done()

	// Stop Accepting new requests
	srv.Shutdown(context.Background())
	// close incoming requests
	close(queue)
	// wait for all workers to finish
	wg.Wait()
	// close results channel
	close(results)
	// wait for aggregator to finish remaning items
	aggWG.Wait()
	log.Print("ShutDown Complete")
}
