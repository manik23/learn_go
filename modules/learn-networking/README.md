# Linux & Networking Systems 🛡️

This module explores the low-level networking stack, focusing on how Go interacts with the Linux kernel (and macOS) to handle high-performance networking.

---

### 🛰️ Progress Tracking
- [x] **Phase 4.1: Socket Observability** (`lsof`, `netstat`)
- [x] **Phase 4.2: TCP State Mastery** (`TIME_WAIT`, `CLOSE_WAIT`, `FIN_WAIT_2`)
- [x] **Phase 4.3: Syscall Barrier** (`dtruss`, `strace`, `EAGAIN` internals)
- [x] **Phase 4.4: TCP Internals** (Window Scaling, MSS, Congestion)
- [ ] **Phase 4.5: The Netpoll Runtime** (Deep dive into Go source: `runtime/netpoll.go`)

---

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

---

### 🛰️ Phase 4.3: The Syscall Barrier

Networking in Go is an abstraction over OS System Calls. Every time you interact with the network, the Go runtime executes a `syscall`.

#### **The "Master" Syscalls**
1.  **`socket()`**: Requests a new file descriptor from the kernel.
2.  **`bind()`**: Anchors the socket to a specific port.
3.  **`listen()`**: Tells the kernel to start queuing incoming connections.
4.  **`accept()`**: Grabs a connection from the queue (returns a new FD).
5.  **`read()` / `write()`**: Moves bytes between User-Space (Go) and Kernel-Space buffers.

#### **Tracing on macOS (`dtruss`)**
Because macOS uses SIP, you often have to trace the binary directly rather than the `go run` wrapper.
```bash
# Build the binary first
go build -o server cmd/server/main.go

# Trace only the networking syscalls
sudo dtruss ./server 2>&1 | grep -E "socket|bind|listen|accept|read|write"
```

> [!NOTE]
> **Tracing on macOS vs. Linux**:
> - **Linux**: Uses `ptrace` and the `strace` utility.
> - **macOS (Darwin)**: Uses `dtrace` and the `dtruss` utility (due to the Xnu kernel architecture).
> - **Why?**: `strace` is built specifically for the Linux kernel's syscall interface. Apple uses the more advanced DTrace engine for system instrumentation.

---

### 🐧 Phase 4.4: Linux Tracing with Docker (`strace`)

Since macOS SIP restricts `dtrace`/`dtruss`, we can pivot to a Linux environment using Docker to see the canonical `strace` output used in production Linux servers.

#### **1. Build and Trace with One Command**
The Makefile now handles cross-compilation (`GOOS=linux GOARCH=arm64`), image building, and running with the correct capabilities.
```bash
make docker-trace-server
```

#### **2. Manual Control**
If you want to explore the Linux environment manually:
```bash
# Build binary and image
make docker-build

# Jump into the shell
make docker-shell
```

#### **4. Detailed `strace` Analysis (The Matrix)**
Based on our live experiment, here is what the kernel was doing:

| Syscall | Observation | Level-Up Insight |
| :--- | :--- | :--- |
| `socket(...) = 4` | Creates a new File Descriptor (4). | The "starting point" for any network communication. |
| `setsockopt(..., SO_REUSEADDR, [1])` | Success! | Allows the server to restart immediately without waiting for `TIME_WAIT` to clear. |
| `accept4(4, ...) = -1 EAGAIN` | **The Masterstroke** | Proves Go uses **Non-blocking I/O**. Instead of hanging, the CPU is freed to run other goroutines. |
| `accept4(4, ...) = 5` | New Connection! | FD 4 remains the "Doorbell", FD 5 becomes the "Private Chat" for this specific client. |
| `read(5, "hello\n", 4096) = 6` | Data In. | You can see the raw bytes moving from the kernel buffer to your Go `[]byte`. |
| `read(5, "", 4096) = 0` | Connection Closed. | In Unix, a read of 0 bytes is the standard "EOF" (End of File/Stream). |

**Sample Log from Experiment:**
```bash
# Server Setup
[pid    12] socket(AF_INET, SOCK_STREAM|SOCK_CLOEXEC|SOCK_NONBLOCK, IPPROTO_TCP) = 4
[pid    12] setsockopt(4, SOL_SOCKET, SO_REUSEADDR, [1], 4) = 0
[pid    12] bind(4, {sa_family=AF_INET6, sin6_port=htons(8080)...}, 28) = 0
[pid    12] listen(4, 4096)             = 0

# The Accept Loop (Non-blocking)
[pid    12] accept4(4, ..., SOCK_CLOEXEC|SOCK_NONBLOCK) = -1 EAGAIN (Resource temporarily unavailable)

# Client Connects
[pid    12] accept4(4, {sa_family=AF_INET6, sin6_port=htons(34784)...}, ...) = 5

# Data Exchange
[pid    12] read(5, "hello\n", 4096)    = 6
[pid    12] write(5, "ECHO: hello\n", 12) = 12

# Client Disconnects
[pid    12] read(5, "", 4096)           = 0
```

---

### 🛠️ Observation Tools (Cross-Platform)

Our `Makefile` now automatically detects your OS and uses the appropriate low-level tools for monitoring connections.

| Feature | macOS Command | Linux Command (`ss`) | Makefile Target |
| :--- | :--- | :--- | :--- |
| **Active Conns** | `lsof -nP -i :8080` | `ss -ntp \| grep 8080` | `make watch-conns` |
| **TCP States** | `netstat -an \| grep 8080` | `ss -ant \| grep 8080` | `make watch-tcp` |
| **TIME_WAIT Count** | `netstat -an \| ... \| wc -l` | `ss -ant \| ... \| wc -l` | `make death-count` |

> [!TIP]
> **Why `ss` over `netstat` on Linux?**
> `ss` (Socket Statistics) is faster and more powerful than the legacy `netstat`. It gets its information directly from the kernel's `tcp_diag` module.

---

### 🛰️ Phase 4.4: TCP Internals (Flow Control & Flow)

In this phase, we look at the internal parameters of the TCP protocol that ensure data is delivered reliably and efficiently.

#### **1. Key Concepts**
- **MSS (Maximum Segment Size)**: The largest amount of data (in bytes) that a device can receive in a single TCP segment. Usually ~1460 bytes for Ethernet.
- **Window Size (`win`)**: The amount of data the receiver is willing to accept without sending an `ACK`. This is how TCP implements **Flow Control**.
- **Retransmission**: If a packet is lost (no `ACK` received), the sender re-sends the data after a timeout.

###### MSS prevents fragmentation, whereas Window Size manages flow control and throughput efficiency.

- **MSS Example (The Slicer)**: 
    If you send a **5,000 byte** image and the MSS is **1,460 bytes** (Standard Ethernet), the TCP layer will automatically slice it into **4 packets** (~1460, ~1460, ~1460, ~620). This prevents the IP layer from having to "fragment" the packets at the router level, which is much more expensive.
- **Window Size Example (The Dam)**:
    If the Server is slow and its internal buffer is filling up, it will send an ACK with `win 1000`. The Client sees this and says, "I can only send 1000 more bytes before I must stop and wait." If the buffer hits 0, it sends **`win 0`** (Zero Window), and the Client completely stops sending data until the Server's app clears the buffer.

#### **2. Observing the Handshake**
Run the following to see how the Client and Server negotiate their "Contract":
```bash
make inspect-tcp
```

#### **3. What to watch for:**
- `options [mss 16344]`: The negotiated maximum packet size for the local loopback.
- `win 65535`: The initial receive window.
- `options [nop,wscale 6]`: Window scaling factor (allows the window to grow beyond 64KB).

#### **4. Detailed Packet Analysis (The Wire)**
From our `tcpdump` experiment, we observed the following flag sequences:

| Flag | Name | Meaning |
| :--- | :--- | :--- |
| **`[S]`** | **SYN** | "Synchronize" - I want to start a connection. |
| **`[S.]`** | **SYN-ACK** | "I hear you, and I also want to connect." |
| **`[.]`** | **ACK** | "Acknowledged" - Handshake complete or data received. |
| **`[P.]`** | **PUSH-ACK** | "Here is data, give it to the app immediately!" |
| **`[F.]`** | **FIN-ACK** | "I'm done sending data. Goodbye." |

> [!IMPORTANT]
> **Observation: The 15-Second Heartbeat**
> In our trace, we saw empty `[.]` packets flowing every 15 seconds:
> `17:13:42 -> 17:13:57 -> 17:14:12`
> This is Go's **TCP Keep-Alive** in action! The kernel sends these "probes" to ensure the other side is still alive, preventing "Ghost Sockets" from hanging around forever if a client disappears without a proper `FIN`.

**Sample Log from Experiment:**
```bash

# TCP Handshake SYN -> SYN-ACK -> ACK
17:15:31 IP6 ::1.61863 > ::1.8080: Flags [S], seq 1659994842, win 65535, options [mss 16324,nop,wscale 6,nop,nop,TS val 3234100448 ecr 0,sackOK,eol], length 0

17:15:31 IP6 ::1.8080 > ::1.61863: Flags [S.], seq 2575719388, ack 1659994843, win 65535, options [mss 16324,nop,wscale 6,nop,nop,TS val 1167892641 ecr 3234100448,sackOK,eol], length 0

17:15:31 IP6 ::1.61863 > ::1.8080: Flags [.], ack 2575719389, win 6371, options [nop,nop,TS val 3234100448 ecr 1167892641], length 0

# Data Exchange (Push + Ack)
17:13:41 IP6 localhost.61694 > localhost.8080: Flags [P.], seq 4229207334:4229207340, ack 629883716, win 6371, length 6
17:13:41 IP6 localhost.8080 > localhost.61694: Flags [.], ack 4229207340, win 6371, length 0

# Keep-Alive Probes (The 15s Heartbeat)
17:13:57 IP6 localhost.8080 > localhost.61694: Flags [.], ack 4229207354, win 6371, length 0
17:13:57 IP6 localhost.61694 > localhost.8080: Flags [.], ack 629883754, win 6370, length 0

# The Final Wave (FIN-ACK)
17:15:09 IP6 localhost.61694 > localhost.8080: Flags [F.], seq 4229207354, ack 629883754, win 6370, length 0
17:15:09 IP6 localhost.8080 > localhost.61694: Flags [.], ack 4229207355, win 6371, length 0
```

---

### 🧠 Phase 4.5: The Netpoll Runtime (The Glue)

This is the internal mechanism that makes Go's networking performant. It prevents networking from being "one thread per connection."

#### **1. The Workflow: From `Read()` to `epoll`**
1.  **Application**: `conn.Read(buf)` 
2.  **Internal**: `poll.FD.Read` calls `syscall.Read`.
3.  **Kernel**: Returns `EAGAIN` (meaning "No data now, try later").
4.  **Runtime**: `poll.FD.WaitRead` is called.
5.  **Park**: `runtime.gopark` puts the goroutine into a `_Gwaiting` state. It stops consuming CPU.
6.  **Poll**: A background `netpoll` thread (or any thread looking for work) calls `epoll_wait` (Linux) or `kevent` (macOS).
7.  **Wake**: When data hits the NIC, the kernel notification system wakes the poller.
8.  **Ready**: `runtime.netpoll` finds the waiting goroutine and calls `runtime.goready(g)`.
9.  **Resume**: The goroutine is put back on the runqueue (`_Grunnable`) and eventually finishes its `Read()` call.

#### **2. Where to look in Go Source**
If you want to read the "Sacred Texts", look for these files in your Go installation:
- `src/runtime/netpoll.go`: The platform-independent logic.
- `src/runtime/netpoll_kqueue.go`: The macOS/BSD implementation.
- `src/runtime/netpoll_epoll.go`: The Linux implementation.
- `src/internal/poll/fd_unix.go`: The bridge between the `net` package and the `runtime`.

> [!IMPORTANT]
> **The Win**: This is why Go can handle **C10k (Target: 10,000 connections)** with only a few MBs of RAM and a handful of OS Threads. The OS only sees a single thread waiting on thousands of sockets using one `epoll_wait` call.

---
