# Cloud Control-Plane Architecture 🧠

This module explores the design of the "Brain" of distributed systems. We focus on how to maintain a consistent global state in a world of partial failures, network retries, and concurrency.

---

### 🛰️ Progress Tracking
- [x] **Phase 5.1: Safety** (Idempotency Keys & Request Caching)
- [x] **Phase 5.2: Stability** (Reconciliation Loops & Level-Triggering)
- [x] **Phase 5.3: Authority** (Leader Election & Leases)
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

### Challenge 3: Chaos Failover (Phase 5.3-B)
**Scenario**: node-1 holds the leader lease but crashes. node-2 and node-3 must detect the failure and elect a new leader.
- **Task**: Run `make cluster-kill-leader` and observe the takeover.
- **Command**: After killing the leader, run `make cluster-logs` and watch for the `[LEASE] Node node-2 acquired lease` log after 15 seconds.
- **Goal**: Zero-downtime authority transition.

### Challenge 4: Sharding (Phase 5.4)
**Scenario**: 10 million resources, 3 nodes. Each node should only reconcile its own "district" to avoid all three scanning the same table.
- **Task**: Implement a `Shard` function using **Consistent Hashing**. Each node picks up only resources where `hash(resourceID) % numNodes == nodeIndex`.
- **Goal**: Horizontal Scale — the brain grows with the cluster.

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

### 3. Remote Integration Tests
```bash
# Run with server active in another terminal
make test-remote
```

### 4. 3-Node Cluster (Leader Election Test)
```bash
# Start all three nodes
make cluster-run

# Watch all logs (with node prefix)
make cluster-logs

# See which nodes are alive
make cluster-status

# Kill node-1 (the leader) to trigger failover
make cluster-kill-leader

# Teardown the cluster
make cluster-stop
```
*After killing the leader, watch the logs. In ~15 seconds (lease TTL), node-2 or node-3 takes over the "Megaphone".* 

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
The `startReconciler` worker runs every 5 seconds. It queries the DB for the current `Observed` count, compares against `Desired`, and creates/deletes records to close the gap. Resources stuck in `PROVISIONING` are automatically completed.

---

## 🏗️ Phase 5.3: Authority (The Megaphone)

Problem: With 3 nodes all running `Reconcile()` against the same DB, they would double-provision and delete-loop forever. A **single source of authority** is required.

### 1. DB Lease (Compare-And-Swap)
A single row in `ControlPlaneLease` table acts as a distributed lock:
- **Row**: `{id: "reconciler-lock", node_id: "node-1", expires_at: now+15s}`
- **Leader Heartbeat**: Every 5s, the leader atomically `UPDATE`s the row, renewing `expires_at`.
- **Follower Check**: Other nodes try the same `UPDATE` but their `WHERE` clause (`node_id = me OR expires_at < now`) only matches if the leader has *died*.
- **Takeover**: When the leader crashes, `expires_at` passes. The first follower to tick wins the `UPDATE` and becomes the new leader.

### 2. Why Atomic UPDATE?
We do NOT use a Read-then-Write pattern:
```go
// ❌ Dangerous: Two nodes can both read "no leader" and both insert
if no_leader { db.Create(lease) }

// ✅ Safe: Only one node's UPDATE can affect a row at a time
db.Where("node_id=me OR expires_at < now").Updates(...)  // RowsAffected == 1 means YOU won
```

### 3. Test It
```bash
make cluster-run   # Start 3 nodes
make cluster-logs  # Watch node-1 leading
make cluster-kill-leader  # Kill node-1 → watch node-2/3 take over
```

---

## ⏭️ What's Next: Phase 5.4 — Scale (Consistent Hashing)

**Problem**: Even with one leader, what if you have 100M resources? One node can't scan 100M rows every 5 seconds.

**Solution**: Divide the work. Each node only reconciles its own "district":
- `node-1` handles `hash(resourceID) % 3 == 0`
- `node-2` handles `hash(resourceID) % 3 == 1`
- `node-3` handles `hash(resourceID) % 3 == 2`

This is **Consistent Hashing** — the foundation of real Kubernetes controllers (each shard-key is a configmap/pod namespace).

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
