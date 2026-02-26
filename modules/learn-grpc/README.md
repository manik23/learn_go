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
