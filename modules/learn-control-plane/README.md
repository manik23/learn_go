# Cloud Control-Plane Architecture 🧠

> **Module 5** of the Senior Go Developer Curriculum. The "Brain" of a distributed system. This module teaches how to maintain consistent global state in a world of partial failures, network retries, and concurrency.

---

## 🛰️ Progress Tracking

- [x] **Phase 5.1: Safety** — Idempotency Keys & Request Caching
- [x] **Phase 5.2: Stability** — Reconciliation Loops & Level-Triggering
- [x] **Phase 5.3: Authority** — Leader Election & DB Leases
- [x] **Phase 5.4: Scale** — Consistent Hashing & Shard-Aware Reconciliation

---

## 🏗️ Strategic Outlook: The Four Pillars

A Control Plane is not an API. It **manages a continuous loop** to ensure reality matches intent:

| Pillar | Mental Model | Core Pattern | Your Bug Without It |
| :--- | :--- | :--- | :--- |
| **Safety** | The Cashed Check | Idempotency Key | Duplicate resources on retry |
| **Stability** | The Thermostat | Level-Triggered Reconciler | Drifted state never self-heals |
| **Authority** | The Megaphone | DB Lease (CAS) | 3-node split-brain corrupts data |
| **Scale** | District Managers | Consistent Hashing | One node scans 100M rows every 5s |

---

## 🛠️ Quick Commands

```bash
# Local single-node mode
make run              # Start server (NODE_ID defaults to "local")
make watch-state      # Live Desired vs Observed view
make scale-up         # Set Desired = 10
make scale-down       # Set Desired = 2
make test             # Unit tests
make test-remote      # Integration tests (server must be running)

# 3-Node cluster (sharding + leader election)
make cluster-run      # Build image, start nodes with NODE_INDEX env
make cluster-logs     # Tail all nodes (prefixed by container name)
make cluster-status   # Health table of running containers
make cluster-kill-leader  # Kill node-1, trigger failover after 15s
make cluster-stop     # Tear down all containers
```

---

## � Best Practices for Building Control Planes

### 1. Always Use the Database as the Source of Truth
Never trust in-memory counters across restarts. On startup, always re-hydrate state from the DB:
```go
// ✅ Correct: read from DB on every boot
db.Model(&ResourceLedger{}).Where("state = ?", PROVISIONED).Count(&p.Observed)

// ❌ Wrong: start from 0 and hope events fill it in
p.Observed = 0
```

### 2. Idempotency is a Two-Layer Problem

| Layer | What it protects | Mechanism |
|---|---|---|
| **Transport** | HTTP retry with same key | `IdempotencyExecution` table — cache StatusCode + Body |
| **Logic** | Business entity created twice | `ResourceLedger` state machine — check before creating |

A client using a *different* key for the same resource bypasses the transport cache but hits the logic guard.

### 3. Use Atomic SQL for Distributed Locks

```go
// ❌ Read-then-Write: Two nodes can both read "no lock" at the same time
if no_lock_found { db.Create(lock) }

// ✅ Atomic UPDATE: Only one node's UPDATE can affect a row at a time
result := db.Where("(node_id = ? OR expires_at < ?)", nodeID, now).Updates(...)
if result.RowsAffected == 0 { /* someone else is leader */ }
```

### 4. Respect Single-Writer Ownership for Shared Fields

Each shared mutable field should have exactly **one function** that writes to it:
```go
// p.Observed is owned by reconcileGlobalState() — only the leader updates it
// reconcileShard() counts its own myObserved locally and NEVER writes to p.Observed
// This prevents a shard-local count from stomping the cluster-wide counter
```

### 5. Separate Global from Local Concerns

```
Reconcile() every tick:
  ├── tryAcquireLease() → if LEADER → reconcileGlobalState()
  │                        (cluster-wide: set Desired, count totals)
  └── reconcileShard()  → ALL nodes, always
                         (per-shard: complete PROVISIONING → PROVISIONED)
```

### 6. Level-Triggered > Edge-Triggered

| Trigger Model | React to | Risk |
|---|---|---|
| Edge-Triggered | State *changes* only | Miss one event → broken state forever |
| **Level-Triggered** ✅ | Current state *vs target* every tick | Self-heals on every iteration |

---

## 🔑 Cheat Sheet: Key Patterns & Code

### Idempotency Middleware (Atomic Cache)
```go
// Wrap the response writer to capture output
bw := &bodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
c.Writer = bw
c.Next()

// Cache if successful
if c.Writer.Status() < 400 {
    db.Create(&IdempotencyExecution{Key: key, StatusCode: c.Writer.Status(), ResponseBody: bw.body.Bytes()})
}
```

### Leader Election (Atomic CAS Lease)
```go
result := db.Model(&ControlPlaneLease{}).
    Where("id = ?", "reconciler-lock").
    Where("(node_id = ? OR expires_at < ?)", nodeID, time.Now()).
    Updates(map[string]interface{}{"node_id": nodeID, "expires_at": time.Now().Add(15 * time.Second)})

isLeader := result.RowsAffected > 0
```

### Consistent Hashing (Shard Ownership)
```go
func (cfg ShardConfig) OwnsShard(resourceID string) bool {
    h := fnv.New32a()
    h.Write([]byte(resourceID))
    return h.Sum32()%uint32(cfg.TotalNodes) == uint32(cfg.NodeIndex)
}
```

### Level-Triggered Reconciler (Observe → Diff → Act)
```go
func (p *Provisioner) reconcileGlobalState(nodeID string) {
    var totalCount, observedCount int64
    db.Model(&ResourceLedger{}).Count(&totalCount)
    db.Model(&ResourceLedger{}).Where("state = ?", PROVISIONED).Count(&observedCount)

    desired := p.getDesired()
    if desired > totalCount {
        // Scale up: create (desired - totalCount) new PROVISIONING stubs
    } else if desired < totalCount {
        // Scale down: delete (totalCount - desired) PROVISIONED records
    }
}
```

### Shard-Aware In-Memory Filtering
```go
// DB doesn't know your hash function — load all, filter in Go
var all []ResourceLedger
db.Find(&all)
for _, r := range all {
    if !shard.OwnsShard(r.ID) { continue }
    // ... process this node's records only
}
```

---

## ⚠️ Anti-Patterns to Avoid

| Anti-Pattern | Why It's Dangerous | Fix |
|---|---|---|
| `if no_lock { Create(lock) }` | Read-then-write race: two nodes both "see no lock" | Use atomic `UPDATE WHERE ... OR expires_at < now` |
| Writing per-shard count into `p.Observed` | Stomps the cluster-wide counter | Keep shard count in a `myObserved` local variable |
| `WHERE hash(id) % 3 == 0` in SQL | DB doesn't know Go's FNV hash | Load all, filter in-memory with `OwnsShard()` |
| Trusting in-memory `p.Desired` on startup | Server restarts with 0 desired | Read from DB on startup |
| Combining Leader gate + Shard gate naively | Followers skip their shard work forever | Separate: leader does global ops, ALL nodes do shard work |

---

## 🧠 Exercises & Challenges

### Challenge 1: The "Atomic Lock" (Phase 5.1-B)
**Scenario**: Two simultaneous requests with the same `X-Idempotency-Key` hit the server. Both check the DB, see nothing, and proceed.
- **Task**: Ensure the DB `UNIQUE` constraint on the key is the last line of defence. Observe the graceful error handling when the second insert fails.

### Challenge 2: Self-Healing Surplus (Phase 5.2-A)
**Scenario**: Scale down from 10 → 3. Verify the reconciler deletes exactly 7 resources.
```bash
make scale-up   # set Desired = 10
# wait for stable
make scale-down # set Desired = 2
make watch-state
```

### Challenge 3: Leader Failover (Phase 5.3-B)
```bash
make cluster-run
make cluster-logs   # verify node-1 is leader
make cluster-kill-leader
# wait 15 seconds... watch node-2 or node-3 take over
```

### Challenge 4: Per-Shard Leases (Phase 5.4 — Advanced)
Replace the single `reconciler-lock` with per-shard locks:
```go
lockKey := fmt.Sprintf("reconciler-lock-shard-%d", shard.NodeIndex)
tryAcquireLease(nodeID, lockKey, db)
```
Now ALL three nodes reconcile in parallel with no single point of coordination.
