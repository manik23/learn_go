# Indexing all Go Concepts

A comprehensive educational repository and reference hub for the Go (Golang) programming language. This project is designed as a multi-module workspace that indexes various Go concepts, ranging from core language features to specific frameworks and advanced interoperability.

## 🎓 Senior Go Developer Curriculum (Mentor Path)

This path focuses on high-performance, resilient, and production-grade Go engineering.

| Module | Topic | Status | Key Learnings |
| :--- | :--- | :--- | :--- |
| 1 | **Production Go Service Patterns** | ✅ Completed | Context propagation, Graceful Shutdown, Connection Pooling, Backpressure. |
| 2 | **Performance Engineering** | ✅ Completed | `pprof` (CPU/Heap), Flame Graphs, Allocation Optimization, `Benchmark` logic vs network overhead. |
| 3 | **gRPC Deeply** | ✅ Completed | Unary/Streaming/Bidi, Interceptors, Protobuf Design, Prometheus Metrics. |
| 4 | **Linux/Networking Systems** | 🏗️ In Progress | TCP States, FD Leaks, `netpoll`, and Syscall Tracing (`strace`). |
| 5 | **Envoy + xDS Control Plane** | 📅 To Do | LDS/RDS/CDS/EDS, Config versioning, Rollout safety. |
| 6 | **Cloud Control-Plane Architecture** | 📅 To Do | Idempotency, Reconciliation loops, Leader election, Sharding. |
| 7 | **DPDK Integration Model** | 📅 To Do | CGO boundaries, Zero-copy interfaces, Memory ownership. |

---

## 📂 Sub-Projects Directory

### 🏆 Advanced Path (Active)
- [🛡️ Production Service Patterns](./modules/prod-service-patterns/) - Graceful shutdown, pooling, and pprof.

### 📚 Fundamentals (Legacy)
- [🚀 Go-Routines](./modules/learn-routines/) - Concurrency patterns and worker pools.
- [🔗 CGO](./modules/learn-cgo/) - Interoperability between Go and C.
- [🌐 GIN Framework](./modules/learn-gin/) - High-performance web development with GORM.
- [🧠 Interview Practice](./modules/go-interview-practise/) - Algorithmic and Go-specific challenges.

---

## 🚀 How to Resume
1. Navigate to `modules/learn-networking`.
2. Resume Phase 4.4: TCP Internals (Window Size & Flow Control).
3. Type `make docker-trace-server` to restart the tracing environment.
