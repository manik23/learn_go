# OS Detection
OS := $(shell uname)

# Variables
PROTO_SRC=proto/service.proto
GEN_OUT=.

.PHONY: generate format explain run-server run-client metrics-grpc metrics-raw docker-build docker-run clean

generate:
	@echo "Generating gRPC code..."
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       $(PROTO_SRC)

format:
	@echo "Formatting code..."
	gofumpt -w ./client/*.go
	gofumpt -w ./server/*.go
	golines -w --max-len=110 ./client/*.go
	golines -w --max-len=110 ./server/*.go

explain:
	@cat README.md | sed -n '/## 🔍 Revision Notes/,$p'

run-server:
	@echo "Starting gRPC server on $(OS)..."
	go run ./server/...

run-client:
	@echo "Starting gRPC client on $(OS)..."
	go run ./client/...

metrics-grpc:
	@echo "Fetching gRPC metrics..."
	curl -s localhost:2112/metrics | grep ^grpc

metrics-raw:
	@echo "Fetching all metrics..."
	curl -s localhost:2112/metrics

# Docker (Linux-based) Parity
docker-build:
	@echo "Building gRPC Linux container..."
	docker build -t learn-grpc-linux .

docker-run: docker-build
	@echo "Running gRPC server in Linux container..."
	docker run -it --rm -p 50051:50051 -p 2112:2112 learn-grpc-linux

clean:
	@echo "Cleaning up generated files..."
	rm -rf proto/*.pb.go
