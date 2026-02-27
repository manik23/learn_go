# Linux & Networking Systems 🛡️

This module explores the low-level networking stack, focusing on how Go interacts with the Linux kernel (and macOS) to handle high-performance networking.

## 🛠️ Experiments & Setup

### 1. TCP Echo Server
A basic TCP server that echoes back any data received. Use this to observe socket lifecycle and syscalls.

```bash
# Start Server
go run cmd/server/main.go

# Run Client
go run cmd/client/main.go
```

### 4. The FD Leak (Process Exhaustion)
Demonstrates the "Too many open files" error.
```bash
LEAK=true go run cmd/server/main.go
```

## 🔍 Revision Notes: Networking Fundamentals

### 1. The Socket Lifecycle
Every network connection in Go follows a kernel-level dance:
- **`socket()`**: Create a file descriptor for communication.
- **`bind()`**: Assign a local address/port to the socket.
- **`listen()`**: Mark the socket as passive (ready to accept).
- **`accept()`**: Block until a client connects (returns a *new* file descriptor for the connection).
- **`connect()`**: (Client-side) Initiate the 3-Way Handshake (`SYN`, `SYN-ACK`, `ACK`).

### 2. The File Descriptor (FD) Problem
In Unix, everything is a file. Every connection is an FD.
- **Limit**: Check your system's limit with `ulimit -n`.
- **Exhaustion**: If your Go app leaks connections, you will get "too many open files" errors.

### 3. Blocking vs. Non-blocking (The `epoll` Magic)
Go's `net` package makes networking look blocking (synchronous), but under the hood:
- **`netpoll`**: Uses `epoll` (Linux) or `kqueue` (macOS).
- **Mechanism**: Instead of one thread per connection, Go registers FDs with the kernel and puts goroutines to sleep. When data arrives, the kernel wakes up the `netpoll` runtime, which then wakes your goroutine.

### 4. The `TIME_WAIT` (Zombie) State
When a TCP connection is closed, the side that initiated the `active close` stays in this state for **2MSL** (Maximum Segment Lifetime), usually 60-120 seconds.

- **Why it exists**:
This state occurs on the side that initiates the connection teardown (the **Active Closer**). It stays in `TIME_WAIT` for **2MSL** (Maximum Segment Lifetime, usually 60-120s) to ensure:

1.  **Reliability**: The final `ACK` is received by the peer. If lost, the peer retransmits `FIN`, and the "zombie" socket is still there to acknowledge it.
2.  **Safety**: Prevents "delayed duplicates" (old packets) from a previous connection from interfering with a new connection using the same Quadruple (Source IP, Source Port, Dest IP, Dest Port).

**The "Bomb" Lesson**: By intentionally creating thousands of connections and closing them immediately, we flood the kernel's connection table with sockets stuck in `TIME_WAIT`, which can prevent new connections from being established (Ephemeral Port Exhaustion).

If you open thousands of short-lived connections to the same destination, you will fill the kernel's connection table. 
- **Symptom**: `dial tcp: assign requested address`.
- **Fix**: Use **Connection Pooling** (Keep-Alives) to reuse existing `ESTABLISHED` sockets.

### 5. TCP 3-Way Handshake
1.  **SYN**: Client sends connection request.
2.  **SYN-ACK**: Server acknowledges and sends its own request.
3.  **ACK**: Client acknowledges. Connection is `ESTABLISHED`.

### 6. File Descriptor (FD) Leak
When a process opens a connection but fails to call `Close()`, that socket remains in the process's **File Descriptor Table**.
- **The Limit**: Systems have hard/soft limits (see `ulimit -n`).
- **The Symptom**: `accept: too many open files`.
- **The Difference**:
    - `TIME_WAIT`: Managed by the **Kernel**. Occurs *after* close.
    - `FD Leak`: Managed by the **Process**. Occurs *instead of* close.

### 7. The Half-Closed Trap: `CLOSE_WAIT` vs `FIN_WAIT_2`
When a side fails to call `Close()` (a leak), the TCP state machine hangs in a semi-closed state:

| State | Responsibility | Description |
| :--- | :--- | :--- |
| **`CLOSE_WAIT`** | **Passive Closer** (The Leak) | The Remote side has closed, the local Kernel has acknowledged (`ACK`), and is now **waiting for the Local Application to call `Close()`**. The File Descriptor remains open. |
| **`FIN_WAIT_2`** | Active Closer (The Hang) | Sent `FIN`, received `ACK`. Now **waiting for the other side to send its `FIN`**. It will stay here until a timeout (usually 60s) because the peer is stuck in `CLOSE_WAIT`. |

- **Senior Gopher Tip**: If you see `CLOSE_WAIT` in production, it's an application bug (missing `Close()`). If you see `FIN_WAIT_2`, you are waiting on a peer that is likely leaking connections. In our example client side is active closer and server side is passive closer. In our case server is leaking connections so it will be in CLOSE_WAIT state and client will be in FIN_WAIT_2 state.


---

### 🛰️ Phase 4.1: Observing Connections

**Tools to use:**
- `lsof -i :8080`: See which processes are holding the port.
- `netstat -an | grep 8080`: Observe connection states (`LISTEN`, `ESTABLISHED`, `TIME_WAIT`).
- `tcpdump -i lo0 port 8080`: Watch the actual packets flowing locally.
