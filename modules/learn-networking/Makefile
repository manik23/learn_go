.PHONY: run-server run-client watch-conns watch-tcp death-count dump-packets trace-server docker-build docker-shell docker-trace-server

# OS Detection
OS := $(shell uname)

ifeq ($(OS), Darwin)
    # macOS
    WATCH_CONNS_CMD = lsof -nP -i :8080
    WATCH_TCP_CMD = netstat -an | grep 8080
    DEATH_COUNT_CMD = netstat -an | grep 8080 | grep TIME_WAIT | wc -l
else
    # Linux (assuming ss is available)
    WATCH_CONNS_CMD = ss -ntp | grep 8080
    WATCH_TCP_CMD = ss -ant | grep 8080
    DEATH_COUNT_CMD = ss -ant | grep 8080 | grep TIME-WAIT | wc -l
endif

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
	@echo "Watching TCP connections on port 8080 ($(OS))..."
	watch -n 1 "$(WATCH_CONNS_CMD)"

watch-tcp:
	@echo "Watching TCP states ($(OS))..."
	watch -n 1 "$(WATCH_TCP_CMD)"

death-count:
	@echo "Counting connections in TIME_WAIT state ($(OS))..."
	watch -n 1 "$(DEATH_COUNT_CMD)"

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
