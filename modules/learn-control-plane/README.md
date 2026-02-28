# Cloud Control-Plane Architecture 🧠

This module explores the design of the "Brain" of distributed systems. We focus on how to maintain a consistent global state in a world of partial failures, network retries, and concurrency.

---

### 🛰️ Progress Tracking
- [x] **Phase 5.1: Safety** (Idempotency Keys & Request Caching)
- [x] **Phase 5.2: Stability** (Reconciliation Loops & Level-Triggering)
- [ ] **Phase 5.3: Authority** (Leader Election & Leases)
- [ ] **Phase 5.4: Scale** (Sharding & Consistent Hashing)

---

## 🏗️ Phase 5: Strategic Outlook

A Control Plane is the **"Brain"** of your system. Unlike a standard API, it doesn't just execute commands; it **manages a loop** to ensure reality matches intent.

### The total outlook for Phase 5:
1.  **Safety First**: Prevent double-spending of resources (Idempotency).
2.  **Autonomous Healing**: Reality drifts; the brain must fix it (Reconciliation).
3.  **Unified Authority**: Ensure only one brain is in charge (Leader Election).
4.  **Distributed Scale**: Spread the load across multiple brains (Sharding).

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

## 📊 Master Cheat Sheet: The Control Plane Mindset

| Pillar | Mental Model | Pattern | Implementation |
| :--- | :--- | :--- | :--- |
| **Safety** | The Cashed Check | Idempotency Key | Middleware intercepting retries. |
| **Stability**| The Thermostat | Reconciliation | Background loop: `Observe -> Diff -> Act`. |
| **Authority**| The Megaphone | Leader Election | CAS (Compare-and-Swap) Lease / Distributed locks (etcd/Consul). |
| **Scale** | District Managers| Sharding | Consistent Hashing. |

---

## 🏗️ Phase 5.2: Reconciliation Loops (The Thermostat)

In Phase 5.2, we moved from purely **Reactive** logic (only acting on API calls) to **Autonomous** logic (self-healing background loops).

### 1. Level-Triggered Design
Unlike "Edge-Triggered" systems that react only when a signal changes, our **Level-Triggered** reconciler constantly compares the current state to the target state.
- **Observe**: Count `PROVISIONED` records in the DB.
- **Analyze**: Calculate `Desired - Observed`.
- **Act**: Create or Delete records to reach equilibrium.

### 2. Implementation: The Loop
The `StartReconciler` worker runs every 5 seconds. If it detects that a resource is in the `PROVISIONING` state for too long (or was created by a previous process that crashed), it completes the work automatically.

---

## 🔍 Revision Notes & Logic Flow

### 1. Why Idempotency? (Safety)
In distributed systems, the network **will** fail. If it fails *after* your server creates a resource but *before* it returns a 200, the client will retry.
- **The Ledger Solution**: Record every unique `X-Idempotency-Key` and its result.
- **The side-effect Guard**: Check the `ResourceLedger` (State machine) for existing entities.

### 2. Why Reconciliation? (Stability)
Manual intervention is for small systems. High-scale systems assume failure.
- **Edge-Triggered**: "Action on change" (Missed signals = broken state).
- **Level-Triggered**: "Action on comparison" (Self-healing on every tick).

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
