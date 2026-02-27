.PHONY: run-server run-client watch-conns watch-packets

run-server:
	@echo "Starting TCP Echo Server..."
	go run cmd/server/main.go

run-client:
	@echo "Starting TCP Client..."
	go run cmd/client/main.go

run-bomb:
	@echo "Launching Port Bomb..."
	go run cmd/bomb/main.go

watch-conns:
	@echo "Watching TCP connections on port 8080..."
	watch -n 1 "lsof -nP -i :8080"

watch-tcp:
	@echo "Watching TCP states with netstat..."
	watch -n 1 "netstat -an | grep 8080"

death-count:
	@echo "Counting connections in TIME_WAIT state..."
	watch -n 1 "netstat -an | grep 8080 | grep TIME_WAIT | wc -l"

dump-packets:
	@echo "Capturing local packets on port 8080 (Requires sudo)..."
	sudo tcpdump -i lo0 -X -vv port 8080

trace-server:
	@echo "Tracing Syscalls for TCP Server (Requires sudo)..."
	sudo dtruss go run cmd/server/main.go 2>&1 | grep -E "socket|bind|listen|accept|read|write"

# Docker (Linux-based) Tracing
# Cross-compile for Linux ARM64 (Assuming user is on Apple Silicon or wants ARM64 Linux)
build-linux:
	@echo "Building Linux ARM64 binary..."
	GOOS=linux GOARCH=arm64 go build -o server-linux cmd/server/main.go

docker-build: build-linux
	@echo "Building Linux-based tracing image..."
	docker build -t learn-networking-linux .

docker-shell:
	@echo "Starting Linux shell (with ptrace capability)..."
	docker run -it --rm --cap-add=SYS_PTRACE --security-opt seccomp=unconfined -v $(PWD):/app learn-networking-linux /bin/bash

# Directly run the container which now has strace as entrypoint
docker-trace-server: docker-build
	@echo "Tracing Syscalls inside Linux Container..."
	docker run -it --rm --cap-add=SYS_PTRACE --security-opt seccomp=unconfined -p 8080:8080 learn-networking-linux
