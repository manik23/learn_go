# Cloud Control-Plane Architecture 🧠

This module explores the design of the "Brain" of distributed systems. We focus on how to maintain a consistent global state in a world of partial failures, network retries, and concurrency.

---

### 🛰️ Progress Tracking
- [x] **Phase 5.1: Idempotency & Side-effects** (Middleware Result Caching + Resource Ledger)
- [ ] **Phase 5.2: Reconciliation Loops** (The Level-Triggered Design)

---

## 🧠 Exercises & Challenges

### Challenge 1: The "Atomic Lock" (Phase 5.1-B)
**Scenario**: Two simultaneous requests with the same `X-Idempotency-Key` hit different CPU cores. Both check the DB, see nothing, and both start the "Expensive Work".
- **Task**: Update `IdempotencyMiddleware` or the DB logic to ensure only **one** request can transition the key from "Absent" to "Executing". 
- **Hint**: Use a database unique constraint or a distributed lock.

### Challenge 2: Self-Healing Surplus (Phase 5.2-A)
**Scenario**: The user updates `Desired` from 10 to 5.
- **Task**: Update `reconciler.go` to detect when `Observed > Desired` and gracefully terminate/delete the extra resources.
- **Goal**: True "Level-Triggered" stability.

### Challenge 3: Chaos & Health (Phase 5.2-B)
**Scenario**: A resource is in the `PROVISIONED` state in the DB, but the actual "service" it represents has crashed.
- **Task**: 
    1. Add a `LastHeartbeat` field to `ResourceLedger`.
    2. Update the `Reconciler` to treat resources with a stale heartbeat (> 30s) as `FAILED`.
    3. Watch the system automatically "replace" the dead resource.
This module adheres to our strict engineering loop:
1. **Idempotency first**: Every write must be safe to retry.
2. **Observability**: Metrics on "Observed State" vs "Desired State".
3. **Chaos-Ready**: We will simulate network partitions to see if our Control Plane can recover.

---

## 🛠️ Stack & Setup
- **Framework**: `Gin` (for efficient routing and middleware architecture)
- **State Management**: In-memory (Transitioning to persistent store soon)

### 1. Run the Control Plane
```bash
make run
```

### 2. Observe the Reconciliation
```bash
make watch-state
```

### 3. Run Remote Integration Tests
Ensure the server is running in another terminal before executing:
```bash
make test-remote
```

---

## 📊 Control-Plane Cheat Sheet

| Concept | Implementation | Analogy |
| :--- | :--- | :--- |
| **Idempotency** | `Idempotency-Key` Ledger | A check to see if we already cashed this specific check. |
| **Reconciliation** | `Desired` - `Observed` | A thermostat trying to reach the target temperature. |
| **Leader Election** | CAS (Compare-and-Swap) Lease | Only one person holds the "Megaphone" at any time. |
| **Sharding** | Rendezvous / Jump Hashing | Distributing the work so no single brain is overloaded. |

---

## 🏗️ Phase 5.1: The Idempotency Ledger

In high-reliability Control Planes, **At-Least-Once Delivery** means retries are inevitable. Idempotency ensures these retries are safe.

### 1. Dual-Layer Protection
We implemented a two-tier deduplication strategy:
1.  **Transport Level (Middleware)**: Caches the `StatusCode` and `ResponseBody` for a specific `X-Idempotency-Key`. If a client retries with the same key, they get an identical response immediately without re-triggering logic.
2.  **Logic Level (Resource Ledger)**: Tracks the state of the business entity (Resource ID). If a different key is used for an existing resource, the logic layer detects the conflict.

### 2. The Implementation Snippet

#### **Middleware: Result Caching**
```go
// Capture the response body
bw := &bodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
c.Writer = bw

c.Next()

// Save result to DB if successful
if c.Writer.Status() < 400 {
    db.Create(&IdempotencyExecution{
        Key: key, 
        StatusCode: c.Writer.Status(), 
        ResponseBody: bw.body.Bytes(),
    })
}
```

#### **Handler: State Guard**
```go
err := p.DB.Where("id = ?", req.ID).First(&ledger).Error
if err == nil {
    if ledger.State == PROVISIONED {
        c.JSON(http.StatusOK, ResourceResponse{...})
        return
    }
}
```

---

## 🏗️ Phase 5.2: Reconciliation Loops (The Thermostat)

**The Goal**: The system should constantly work to make the **Observed State** (reality) match the **Desired State** (user's intent).

### Core Concepts:
1. **Edge-Triggered**: Act only when a signal changes (Efficient but risky if signals are lost).
2. **Level-Triggered**: Act by comparing state at every interval (Self-healing and robust).
3. **The Loop**: `Observe` -> `Analyze` (Diff) -> `Act`.


## 🔍 Revision Notes: The Idempotency Key
In a distributed system, **At-Least-Once Delivery** is common. 
- If the network fails *after* the server succeeds but *before* the client gets the ACK, the client will retry.
- Without a ledger, you would create two resources. 
- **The Fix**: The client generates a unique `x-idempotency-key`. The server records this key in a durable store (Redis/DB) before committing the side-effect.
