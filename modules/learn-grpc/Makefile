.PHONY: generate run-server run-client clean

# Variables
PROTO_SRC=proto/service.proto
GEN_OUT=.

generate:
	@echo "Generating gRPC code..."
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       $(PROTO_SRC)

run-server:
	@echo "Starting gRPC server..."
	go run server/main.go

run-client:
	@echo "Starting gRPC client..."
	go run client/main.go

clean:
	@echo "Cleaning up generated files..."
	rm -rf proto/*.pb.go
