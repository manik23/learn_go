# Indexing all Go Concepts

A comprehensive educational repository and reference hub for the Go (Golang) programming language. This project is designed as a multi-module workspace that indexes various Go concepts, ranging from core language features to specific frameworks and advanced interoperability.

## 🎓 Senior Go Developer Curriculum (Mentor Path)

This path focuses on high-performance, resilient, and production-grade Go engineering.

| Module | Topic | Status | Key Learnings |
| :--- | :--- | :--- | :--- |
| 1 | **Production Go Service Patterns** | ✅ Completed | Context propagation, Graceful Shutdown, Connection Pooling, Backpressure. |
| 2 | **Performance Engineering** | ✅ Completed | `pprof` (CPU/Heap), Flame Graphs, Allocation Optimization, `Benchmark` logic vs network overhead. |
| 3 | **gRPC Deeply** | ✅ Completed | Unary/Streaming/Bidi, Interceptors, Protobuf Design, Prometheus Metrics. |
| 4 | **Linux/Networking Systems** | ✅ Completed | TCP States, FD Leaks, `netpoll`, Syscall Tracing (`strace`), Packet Flags. |
| 5 | **Cloud Control-Plane Architecture** | 🚧 In Progress | Idempotency ✅, Reconciliation ✅, Leader Election ✅, Sharding 📅 |
| 6 | **Envoy + xDS Control Plane** | 📅 To Do | LDS/RDS/CDS/EDS, Config versioning, Rollout safety. |
| 7 | **DPDK Integration Model** | 📅 To Do | CGO boundaries, Zero-copy interfaces, Memory ownership. |

---

## 📂 Sub-Projects Directory

### 🏆 Advanced Path (Active)
- [🛡️ Production Service Patterns](./modules/prod-service-patterns/) - Graceful shutdown, pooling, and pprof.
- [🛡️ gRPC Deeply](./modules/learn-grpc/) - Unary/Streaming/Bidi, Interceptors, Protobuf Design, Prometheus Metrics.
- [🌐 Linux/Networking Systems](./modules/learn-networking/) - TCP States, FD Leaks, `netpoll`, and Syscall Tracing (`strace`).

### 📚 Fundamentals (Legacy)
- [🚀 Go-Routines](./modules/learn-routines/) - Concurrency patterns and worker pools.
- [🔗 CGO](./modules/learn-cgo/) - Interoperability between Go and C.
- [🌐 GIN Framework](./modules/learn-gin/) - High-performance web development with GORM.
- [🧠 Interview Practice](./modules/go-interview-practise/) - Algorithmic and Go-specific challenges.

---

## 🏗️ The Senior Learning Workflow
For every module in this curriculum, we adhere to a strict **Systemic Engineering Loop**:
1.  **Observability First**: Every module must have a `Makefile` with targets for running, watching, and tracing.
2.  **Revision Logs**: All conceptual learnings and "Senior Side-bars" are documented in the local `README.md`.
3.  **Kernel-Space Equality**: If a tool is restricted by macOS (SIP), we pivot immediately to a **Linux-based Docker Container** to see the canonical behavior.
4.  **Hands-on Chaos**: We don't just write code; we break it (FD leaks, Port bombs) to understand failure modes.
5.  **Cheat Sheets**: Every module ends with a "Revision Cheat Sheet" containing the most critical commands and codes for quick reference.

## 🚀 How to Resume
1.  **Objective**: Start **Module 5: Cloud Control-Plane Architecture**.
2.  **Initial Task**: Initialize the `learn-control-plane` module and explore **Idempotency** vs **Side-effects**.
3.  **Core Concepts**: Reconciliation loops, Leader Election (etcd/consul), and Sharding strategies.
