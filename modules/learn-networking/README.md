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
    1.  **Reliability**: Ensures the final `ACK` is received by the peer. If lost, the peer retransmits `FIN`, and the "zombie" socket is still there to acknowledge it.
    2.  **Safety**: Prevents "delayed duplicates" (old packets) from a previous connection from interfering with a new connection using the same Quadruple (Source IP, Source Port, Dest IP, Dest Port).
- **The "Port Bomb" Lesson**: 
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
When a connection is "leaked" (one side fails to call `Close()`), the TCP state machine gets stuck:

| State | Responsibility | Meaning |
| :--- | :--- | :--- |
| **`FIN_WAIT_2`** | Active Closer | Sent `FIN`, received `ACK`. Waiting for the other side to send its `FIN`. |
| **`CLOSE_WAIT`** | **Passive Closer** | Received `FIN`, sent `ACK`. **Waiting for the local application to call `Close()`**. |

- **Why this is dangerous**: If your app is stuck in `CLOSE_WAIT`, it means you are leaking File Descriptors. The kernel will keep that socket open forever (or for a very long timeout), eventually leading to `Too many open files`.
- **The Culprit**: Usually a missing `defer conn.Close()` or an error path that returns early without cleaning up.

---

### 🛰️ Phase 4.1: Observing Connections

**Tools to use:**
- `lsof -i :8080`: See which processes are holding the port.
- `netstat -an | grep 8080`: Observe connection states (`LISTEN`, `ESTABLISHED`, `TIME_WAIT`).
- `tcpdump -i lo0 port 8080`: Watch the actual packets flowing locally.
