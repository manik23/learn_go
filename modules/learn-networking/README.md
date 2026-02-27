# Linux & Networking Systems 🛡️

This module explores the low-level networking stack, focusing on how Go interacts with the Linux kernel (and macOS) to handle high-performance networking.

---

## 🛠️ Experiments & Setup

Run these experiments to see the networking stack in action.

### 1. TCP Echo Server
A basic TCP server that echoes back any data received.
```bash
# Terminal 1: Start Server
go run cmd/server/main.go

# Terminal 2: Run Client
go run cmd/client/main.go
```

### 2. The Port Bomb (Ephemeral Port Exhaustion)
Demonstrates how fast the kernel fills up with `TIME_WAIT` sockets.
```bash
# Terminal 1: Ensure server is running
# Terminal 2: Launch the bomb
go run cmd/bomb/main.go

# Terminal 3: Observe the counts
netstat -an | grep 8080 | grep TIME_WAIT | wc -l
```

### 3. The FD Leak (Process Exhaustion)
Demonstrates the "Too many open files" error caused by application-level leaking.
```bash
# Terminal 1: Start server in leak mode
LEAK=true go run cmd/server/main.go

# Terminal 2: Run the bomb to hit the limit
go run cmd/bomb/main.go
```

### 4. The Multiplexing Demo (`netpoll` in action)
Simulates hundreds of concurrent connections to show how Go handles high concurrency without thread-per-connection overhead.
```bash
go run cmd/multiplex/main.go
```

---

## 🔍 Revision Notes: Networking Fundamentals

### 1. The Socket Lifecycle
Every network connection in Go follows a kernel-level dance:
- **`socket()`**: Create a file descriptor (FD).
- **`bind()`**: Assign a local address/port to the socket.
- **`listen()`**: Mark the socket as passive (ready to accept).
- **`accept()`**: Wait for a client (returns a *new* FD for the specific connection).
- **`connect()`**: (Client-side) Initiate the 3-Way Handshake.

### 2. The File Descriptor (FD) Problem
In Unix, **everything is a file**. Every connection is an FD in the process table.
- **Limit**: Check your system limit with `ulimit -n`.
- **Exhaustion**: If your Go app leaks connections (fails to call `Close()`), you will get `accept: too many open files`.

### 3. Blocking vs. Non-blocking (The `netpoll` Magic)
To a Go programmer, `conn.Read()` looks **synchronous**. However, Go uses **Non-blocking I/O** via `netpoll` (epoll/kqueue).

- **The G-M-P Connection**:
    - When a Goroutine (G) would block on I/O, the Go runtime **parks** it (`_Gwaiting`).
    - The OS Thread (M) is freed to run another G.
    - The `netpoll` system registers the FD with the kernel poller.
    - When data arrives, the kernel wakes the poller, and Go puts the G back into the **runnable** queue (`_Grunnable`).
- **The Win**: This allows thousands of concurrent connections (C10k Problem) with only a few OS threads.

### 4. TCP State Mastery

#### **The `TIME_WAIT` (Zombie) State**
Occurs on the side that initiates the `active close`. It stays for **2MSL** (60-120s) to ensure:

By intentionally creating thousands of connections and closing them immediately, we flood the kernel's connection table with sockets stuck in `TIME_WAIT`, which can prevent new connections from being established (Ephemeral Port Exhaustion).

1.  **Reliability**: The final `ACK` is received. If lost, the peer retransmits `FIN`.
2.  **Safety**: Old "delayed duplicate" packets don't interfere with new connections  using the same Quadruple (Source IP, Source Port, Dest IP, Dest Port).
> [!WARNING]
> Flooding short-lived connections causes **Ephemeral Port Exhaustion**.


If you open thousands of short-lived connections to the same destination, you will fill the kernel's connection table. 
- **Symptom**: `dial tcp: assign requested address`.
> **Fix**: Use **Connection Pooling** to reuse `ESTABLISHED` sockets.

#### **TCP 3-Way Handshake**
1.  **SYN**: Client sends connection request.
2.  **SYN-ACK**: Server acknowledges and sends its own request.
3.  **ACK**: Client acknowledges. Connection is `ESTABLISHED`.

#### **File Descriptor (FD) Leak**
When a process opens a connection but fails to call `Close()`, that socket remains in the process's **File Descriptor Table**.
- **The Limit**: Systems have hard/soft limits (see `ulimit -n`).
- **The Symptom**: `accept: too many open files`.
- **The Difference**:
    - `TIME_WAIT`: Managed by the **Kernel**. Occurs *after* close.
    - `FD Leak`: Managed by the **Process**. Occurs *instead of* close.

#### **The Half-Closed Trap: `CLOSE_WAIT` vs `FIN_WAIT_2`**
Occurs when one side fails to call `Close()` (Leaking).
When a side fails to call `Close()` (a leak), the TCP state machine hangs in a semi-closed state:

| State | Responsibility | Description |
| :--- | :--- | :--- |
| **`CLOSE_WAIT`** | **Passive Closer** (The Leak) | Received `FIN`, sent `ACK`. **Waiting for the local application code to call `Close()`**. |
| **`FIN_WAIT_2`** | Active Closer (The Hang) | Sent `FIN`, received `ACK`. Now **waiting for the peer to send its `FIN`**. |

> [!TIP]
> - `CLOSE_WAIT` = **Local** App Bug (You forgot `Close()`).
> - `FIN_WAIT_2` = **Peer** App Bug (They forgot `Close()`).
> 
> *Experiment Note*: In our `LEAK=true` test, the **Server** stays in `CLOSE_WAIT` because it never calls `Close()`. The **Client** stays in `FIN_WAIT_2` waiting for that final `FIN`.

---

### 🛰️ Phase 4.1: Observing Connections

**Tools to use:**
- `lsof -nP -i :8080`: See which processes hold the FDs.
- `netstat -an | grep 8080`: Observe connection states.
- `tcpdump -i lo0 port 8080`: Watch raw packets (SYN, ACK, FIN) flowing locally.
- `ulimit -n`: Check process file descriptor limits.
