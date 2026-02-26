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

## 🔍 Implementation Patterns
1. **Unary RPC**: A simple request-response pattern.
2. **Server Streaming**: A single request followed by multiple response messages.
3. **Context & Deadlines**: (In Progress) Implementing client-side timeouts and server-side cancellation checks.
4. **Rich Errors**: (In Progress) Using `google.golang.org/grpc/status` for semantic error codes.
