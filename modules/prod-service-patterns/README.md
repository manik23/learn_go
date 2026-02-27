# Production Go Service Patterns 🛡️

This module focuses on the patterns required to build resilient, observable, and high-performance services in Go. We move beyond "it works" to "it survives production."

---

### 🛰️ Progress Tracking
- [x] **Graceful Shutdown**: Handling `SIGINT`/`SIGTERM` without dropping active requests.
- [x] **Context Propagation**: Slicing and dicing timeouts through the execution chain.
- [x] **Backpressure (Semaphores)**: Using channels to limit resource-heavy operations (e.g., DB writes).
- [x] **Observability (pprof)**: Real-time profiling of CPU and Heap allocations.
- [x] **Load Testing & Chaos**: Observing service behavior under extreme pressure.

---

## 🏗️ Core Patterns Explained

### 1. Graceful Shutdown
**The Goal**: Never terminate a process while a customer's request is still being processed.
- **How**: We use `signal.NotifyContext` to listen for OS signals. When a signal arrives, the context is cancelled, triggering a controlled shutdown of all components.

```go
// Inside main.go
ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer cancel()

// Listen for shutdown signal
<-ctx.Done()
log.Println("Shutting down server...")

// Give active requests 5 seconds to finish
shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
if err := server.Shutdown(shutdownCtx); err != nil {
    log.Fatalf("Server forced to shutdown: %v", err)
}
```

### 2. Context & Timeouts
**The Goal**: Prevent "zombie" requests that eat up resources after a client has already disconnected.
- **How**: We use a tree of contexts. If the parent (Request) times out, all children (Auth, Validate, DB) are notified immediately.

```go
func (app *App) handleProcess(w http.ResponseWriter, r *http.Request) {
    // 1. Create a derived context with a 2-second timeout
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()

    // 2. Pass this ctx down to every function
    if err := app.stepAuth(ctx); err != nil {
        // Handle timeout vs application error
    }
}
```

### 3. Backpressure (Semaphores)
**The Goal**: Protecting fragile resources (like a Database) from being overwhelmed.
- **How**: A buffered channel acts as a "Ticket Office." If there are no tickets left, the worker must wait, preventing a "Thundering Herd."

```go
// The "Token" channel (Semaphore)
Token := make(chan struct{}, CAPACITY)

func (s *Server) stepStore(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-s.Token: // Acquire token
        defer func() { s.Token <- struct{}{} }() // Release token
        return s.db.Save(ctx, data)
    }
}
```

---

## 🛠️ Experiments & Setup

### 1. Run the Service
```bash
make run
```

### 2. Observe Memory & CPU (pprof)
The service exposes a pprof server on Port 9000.
```bash
# Capture and view CPU profile
make profile-cpu

# View Heap allocations
make profile-heap
```

### 4. Load Testing & Chaos (Hands-on)
Observe how the service handles failure paths using environment variables.

#### **Experiment: Slow Authentication (Timeout test)**
```bash
# Start server with 3s delay in Auth
SLOW_AUTH=true make run

# Trigger a request in another terminal
curl http://localhost:8080/process
# Output should be: gateway timeout (context deadline exceeded)
```

#### **Experiment: Database Failure**
```bash
# Start server with forced DB errors
DB_FAILURE=true make run

# Trigger a request
curl http://localhost:8080/process
# Output should be: Internal Server Error (database connection refused)
```

---

## 🔍 Revision Notes: Senior Insights

### 1. The "Select" Pattern for Cancellation
In `stepStore`, notice we select on three things:
1. `ctx.Done()`: User cancelled or timed out.
2. `appConfig.ctx.Done()`: The server itself is shutting down.
3. `appConfig.DB.Token`: The database is ready to accept work.
This is the correct way to handle multi-layered cancellation in Go.

### 2. PProf in Production
Never expose `pprof` on a public port. Always bind it to `127.0.0.1` or a private management network. It is the "MRI machine" for your Go service.

## 📊 PProf Cheat Sheet (The Senior MRI)

When your service is slow or crashing, use these commands to diagnose the bottleneck.

| Goal | Command | Analysis Target |
| :--- | :--- | :--- |
| **CPU Spikes** | `go tool pprof http://localhost:9000/debug/pprof/profile?seconds=30` | Find which function is hogging the CPU. |
| **Memory Leaks** | `go tool pprof http://localhost:9000/debug/pprof/heap` | Identify objects that aren't being Garbage Collected. |
| **Deadlocks** | `go tool pprof http://localhost:9000/debug/pprof/block` | See where goroutines are stuck waiting on mutexes/channels. |
| **Goroutine Blast** | `curl http://localhost:9000/debug/pprof/goroutine?debug=1` | Count and list all active goroutines. |

### Visualizing the Data
You can use the `-http` flag to open a web-based UI with Flame Graphs:
```bash
go tool pprof -http=:8081 http://localhost:9000/debug/pprof/heap
```
Flame Graphs help you identify the "Hot Path"—the execution sequence that consumes the most resources.

---

### 🐧 Linux/Docker Parity
If you want to see how these patterns behave on a Linux kernel with different scheduler behavior:
```bash
make docker-run
```
