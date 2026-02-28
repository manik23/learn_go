# Cloud Control-Plane Architecture 🧠

This module explores the design of the "Brain" of distributed systems. We focus on how to maintain a consistent global state in a world of partial failures, network retries, and concurrency.

---

### 🛰️ Progress Tracking
- [ ] **Phase 5.1: Idempotency & Side-effects** (The `Idempotency-Key` Pattern)
- [ ] **Phase 5.2: Reconciliation Loops** (The Level-Triggered Design)
- [ ] **Phase 5.3: Leader Election** (Distributed Locking with etcd/consul)
- [ ] **Phase 5.4: Sharding & Work Distribution** (Consistent Hashing)

---

## 🏗️ The Senior Learning Workflow
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

## 🔍 Revision Notes: The Idempotency Key
In a distributed system, **At-Least-Once Delivery** is common. 
- If the network fails *after* the server succeeds but *before* the client gets the ACK, the client will retry.
- Without a ledger, you would create two resources. 
- **The Fix**: The client generates a unique `x-idempotency-key`. The server records this key in a durable store (Redis/DB) before committing the side-effect.
