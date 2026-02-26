# gRPC Learning Module 🛰️

This module covers the implementation of gRPC services in Go, focusing on Unary and Server-Side Streaming patterns, as well as production-grade error handling and deadlines.

## 🛠️ Scaffolding & Setup

To recreate this module from scratch, follow these steps:

### 1. Initialize the Module
```bash
mkdir -p modules/learn-grpc/proto
cd modules/learn-grpc
go mod init learn-grpc
```

### 2. Install Dependencies
You need the Protocol Buffer compiler (`protoc`) and the Go plugins.
```bash
# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Ensure your GOPATH/bin is in your PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

### 3. Generate Code
The `Makefile` simplifies the generation process.
```bash
make generate
```
*This runs: `protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/service.proto`*

### 4. Project Structure
- `proto/`: Contains the `.proto` definition and generated `.pb.go` files.
- `server/`: Implementation of the gRPC server.
- `client/`: Implementation of the gRPC client.
- `Makefile`: Automation for generation and running.

## 🚀 How to Run

### Start the Server
```bash
make run-server
```

### Run the Client
```bash
make run-client
```

## 🔍 Revision Notes: gRPC Interceptors

### 1. Interceptor Analysis (Unary)
The Unary Interceptor (`grpc.UnaryServerInterceptor`) has a straightforward signature:
- **`ctx`**: Standard context for deadlines and metadata.
- **`req`**: The **fully unmarshaled** request message (e.g., `*pb.HelloRequest`). It is available as a parameter because Unary calls have exactly one request.
- **`info`**: Contains metadata about the call (like `FullMethod`).
- **`handler`**: The next step in the pipeline. Calling `handler(ctx, req)` executes your service logic.

### 2. Why Stream Interceptors are "Hard"
The Stream Interceptor (`grpc.StreamServerInterceptor`) is significantly different:
- **The `srv` Parameter**: Unlike Unary, the `srv` parameter here is the **Service Implementation** (your `*server` struct), NOT the request message.
- **The Problem**: You cannot cast `srv` to a request type (e.g., `srv.(*pb.HelloRequest)` will fail). This is because a stream can have zero or many messages; the gRPC runtime doesn't know which one you want to validate yet.
- **Solution**: To validate message content in a stream, you must "wrap" the `ServerStream` and intercept the `RecvMsg` call, which is complex and boilerplate-heavy.

### 3. How Metadata (Headers) Helps
Instead of looking inside the message body, we use **gRPC Metadata** (HTTP/2 Headers):
- **Universal Access**: Metadata is stored in the `context`, which is available in **both** Unary and Stream interceptors before the first message is even read.
- **Efficiency**: The server can reject a request (e.g., "Invalid Version") by just looking at the headers, without wasting CPU cycles unmarshaling the message body.
- **Clean Contracts**: Business logic (the `.proto` file) stays focused on data, while meta-logic (Auth, Versioning, Tracing) lives in the transport layer.

## 🛡️ Protobuf Best Practices

### 1. Field Numbers (Tags) & Payload Size
- **Tag 1-15**: Take only **1 byte** for the tag + wire type. Use these for your most frequently used (hot) fields.
- **Tag 16-2047**: Take **2 bytes**.
- **Ordering**: Numbers are encoded sequentially. Grouping logic together improves readability, but number choice dictates size.

### 2. Backward Compatibility
- **Never Change Tag Numbers**: This is the cardinal sin of Protobuf. It will break every existing client/server.
- **Reserved Keyword**: When deleting a field, always reserve the tag and name. This prevents future developers from reusing them and causing silent data corruption.
  ```proto
  message User {
    reserved 2, 5 to 8;
    reserved "old_email", "temp_token";
  }
  ```
- **Deprecation**: Use `[deprecated = true]` to signify that a field should no longer be used while keeping it wire-compatible.

### 3. Type Selection
- **Varints**: `int32`, `int64`, `uint32`, `uint64` use variable-length encoding. Small numbers = small size.
- **Fixed Size**: If values are frequently large (e.g., hashes, IDs), use `fixed32` or `fixed64`. They are always 4 or 8 bytes, which is sometimes more efficient than a large varint.
- **Enums**: Always define a `0` value as the default (e.g., `STATE_UNKNOWN = 0;`).

### 4. Code Generation
- Ensure you use `paths=source_relative` to maintain sane import paths in large Go projects.

### 5. Protobuf Keywords & Type Reference

### 5. Protobuf Keywords & Type Reference
| Keyword | Purpose | Example |
| :--- | :--- | :--- |
| `syntax` | Defines the proto version. | `syntax = "proto3";` |
| `package` | Prevents name clashes between projects. | `package orders.v1;` |
| `import` | Use definitions from other proto files. | `import "google/protobuf/any.proto";` |
| `option` | Configures the generator. | `option go_package = "./proto";` |
| `repeated` | Defines a list/array of items. | `repeated string tags = 1;` |
| `map<K,V>` | Defines a key-value pair map. | `map<string, int32> scores = 1;` |
| `oneof` | Mutually exclusive fields (Memory efficient). | `oneof contact { string email = 1; string phone = 2; }` |
| `enum` | A set of named constants. | `enum Status { UNKNOWN = 0; ACTIVE = 1; }` |

#### **Example: Advanced Features**
```proto
message UserProfile {
  // Use 'reserved' to block old tags/names after deletion
  reserved 3, 4;
  reserved "old_title";

  string username = 1;
  
  // Deprecated field: still functional but warns developers
  string legacy_id = 2 [deprecated = true];

  // Map type
  map<string, string> metadata = 5;

  // List of items
  repeated string roles = 6;

  // Mutually exclusive: only one of these will be set
  oneof authentication {
    string oauth_token = 7;
    string password_hash = 8;
  }

  enum Role {
    ROLE_UNSPECIFIED = 0; // Essential default
    ROLE_ADMIN = 1;
    ROLE_USER = 2;
  }
}
```

## 📡 Advanced Observability & Metadata Patterns

### 1. Interceptor Context Augmentation
In Unary interceptors, you can generate a Request ID and inject it into the context so that all downstream logic (handlers, database calls, etc.) can use it for tracing.

**The "Incoming Context" Trick**:
Instead of just using `context.WithValue`, you can modify the **Incoming Metadata**. This allows any downstream logic using `metadata.FromIncomingContext` to see the generated ID as if it was sent by the client.
```go
func AddIDToCtx(ctx context.Context) context.Context {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok { md = metadata.New(nil) }

    if ids := md.Get("x-request-id"); len(ids) == 0 {
        md.Set("x-request-id", uuid.New().String())
    }
    // Re-wrap the context with the modified metadata
    return metadata.NewIncomingContext(ctx, md)
}

```

### 2. Incoming vs. Outgoing Metadata
- **Incoming Metadata**: Headers sent by the client to YOU. Use `metadata.FromIncomingContext`.
- **Outgoing Metadata**: Headers YOU send to a downstream service (if this server acts as a client). Use `metadata.AppendToOutgoingContext`.
- **Common Gotcha**: Appending to *Outgoing* context does not make the data visible to your own local handlers.

### 3. The "Immutability" of Streams
In a `StreamServerInterceptor`, you cannot simply update a `ctx` variable and expect the handler to see it. The `grpc.ServerStream` object carries its own context that is established before the interceptor is called.
- **The Problem**: A stream is a long-lived connection. The `grpc.ServerStream` object is initialized with a context that is effectively "locked." Even if you update a local `ctx` variable in your interceptor, the `handler` won't see it because it pulls the context directly from the `stream` object.
- **The Fix (Stream Wrapper)**: You must create a struct that wraps the original stream and overrides the `Context()` method. This allows you to "hijack" the stream and provide a context enriched with your Request IDs or other metadata.

```go
type wrappedStream struct {
    grpc.ServerStream
    ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
    return w.ctx
}

// In Interceptor:
newCtx := AddIDToCtx(originalCtx)
wrapped := &wrappedStream{ServerStream: originalStream, ctx: newCtx}
return handler(srv, wrapped)
```

### 4. The "Safety Net": Recovery Interceptor
In a production Go service, a single `panic` in a handler should never bring down the entire server. 
- **Pattern**: Always wrap your interceptor logic (especially those that start chains) in a `defer recover()` block.
- **Enhanced Safety**: Use `grpc.ChainUnaryInterceptor` and `grpc.ChainStreamInterceptor` to ensure a dedicated Recovery interceptor is the **first** line of defense.

### 5. 📊 Prometheus: The Observability Standard

Prometheus is an open-source systems monitoring and alerting toolkit. It is the industry standard for cloud-native observability because it uses a **Pull Model** (the server scrapes the targets) and a powerful query language (**PromQL**).

#### **Core Metric Types**

| Type | Purpose | Behavior |
| :--- | :--- | :--- |
| **Counter** | Track totals | Only goes **up** (e.g., `requests_total`). Resets to 0 only on restart. |
| **Gauge** | Track current state | Can go **up and down** (e.g., `current_memory_usage`, `active_goroutines`). |
| **Histogram** | Track distributions | Samples observations (e.g., request duration) and counts them in configurable buckets. |
| **Summary** | Track quantiles | Similar to histogram, but calculates configurable quantiles (e.g., p95, p99) over a sliding time window. |


| Type | Purpose | Behavior | Example Code |
| :--- | :--- | :--- | :--- |
| **Counter** | Track totals | Goes **up** only. | `prometheus.NewCounter(...)` |
| **Gauge** | Track state | Goes **up and down**. | `prometheus.NewGauge(...)` |
| **Histogram** | Distributions | Counts into buckets. | `prometheus.NewHistogram(...)` |

```go
var (
    // Count events: How many times has this happened?
    tasksCompleted = prometheus.NewCounter(prometheus.CounterOpts{Name: "tasks_completed_total"})

    // Current state: How many users are online RIGHT NOW?
    activeConnections = prometheus.NewGauge(prometheus.GaugeOpts{Name: "active_connections"})

    // Latency distribution: How long do requests take (with buckets)?
    requestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
        Name:    "request_duration_seconds",
        Buckets: prometheus.DefBuckets,
    })
)
```

#### **Labels: The Power of Dimensions**
Labels allow you to attach metadata to a single metric name, turning one number into a multi-dimensional matrix.
- **Example**: `learn_grpc_greetings_total{method="SayHello", version="1.0.0"}`

**Real-world Query (PromQL): Calculate p99 Latency per Method**
Labels enable you to calculate complex statistics across different dimensions. To find the 99th percentile latency of all `OK` requests over the last 5 minutes:
```promql
histogram_quantile(0.99, 
  sum by (le, method) (
    rate(grpc_server_handling_seconds_bucket{status="OK"}[5m])
  )
)
```
*This query tells you the latency threshold that 99% of your "Healthy" requests fall under, grouped by the RPC method.*

#### **⚠️ Senior Warning: Cardinality Explosion**
Every unique combination of labels creates a new **time series** in Prometheus.
- **Safe**: `method`, `status_code`, `client_version` (Low cardinality, < 100 values).
- **Dangerous**: `user_id`, `request_id`, `email` (High cardinality, millions of values).
- **Consequence**: High cardinality will consume all RAM on your Prometheus server and crash the monitoring infrastructure.

#### **Implementation Strategy**
In this project, we used two layers of observability:
1.  **Automated**: Using `go-grpc-prometheus` interceptors to catch standard gRPC metrics (Latency, Status Codes).
2.  **Custom**: Defining a `totalGreetings` vector in `metrics.go` to track business events with specific labels like `client_version`.

---

## 🏆 Key Takeaways for Senior Review

1.  **Semantics Over Payload**: Always use gRPC Metadata for infrastructure concerns (Auth, Tracing, Versioning). Keep the `.proto` messages strictly for business data.
2.  **Order Matters in Chaining**: In `grpc.ChainUnaryInterceptor`, the order of execution is Top-to-Bottom. Always place **Recovery** first (to catch panics), then **Auth/Validation**, then **Logging/Observability**.
3.  **The Stream Wrapper Necessity**: If you want to propagate metadata into a `StreamServer` method, a simple `context.WithValue` is not enough. You **must** wrap the stream interface to override the behavior of `stream.Context()`.
4.  **The "Transparent" Trace**: Modifying the **Incoming Context** metadata inside an interceptor is a powerful way to ensure the entire execution tree (even 3rd party libs) can find the `request_id` using standard gRPC methods.
4.  **Value vs. Header**: Use `Metadata` for things that need to cross the network (ID, Version). Use `context.WithValue` only for things that are local to the current process/memory.
5.  **Defensive Streaming**: Because streams are long-lived, always check `stream.Context().Err()` inside your loops to avoid "Zombie Streams" that waste resources after a client disconnects.
