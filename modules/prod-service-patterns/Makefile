# OS Detection
OS := $(shell uname)

# Parameters
PORT=8080
PPROF_PORT=9000
SERVER_URL=http://localhost:$(PORT)/process
PPROF_URL=http://localhost:$(PPROF_PORT)/debug/pprof

.PHONY: run bench bench-remote profile-cpu profile-heap clean docker-build docker-run

run:
	@echo "Starting production service on :$(PORT) ($(OS))..."
	@echo "pprof enabled on :$(PPROF_PORT)..."
	go run main.go

bench:
	@echo "Running local benchmarks (logic only)..."
	go test -v -bench=BenchmarkHandleProcessLocal -benchmem -run=^$$ . -benchtime=10s

bench-remote:
	@echo "Running remote benchmarks (network + logic)..."
	@echo "Make sure the server is running in another terminal!"
	go test -v -bench=BenchmarkHandleProcessRemote -benchmem -run=^$$ . -benchtime=10s

profile-cpu:
	@echo "Capturing 30s CPU profile..."
	go tool pprof -http=:8001 "$(PPROF_URL)/profile?seconds=30"

profile-heap:
	@echo "Capturing current heap profile..."
	go tool pprof -http=:8002 "$(PPROF_URL)/heap"

# Docker (Linux-based) Parity
docker-build:
	@echo "Building Linux-based production container..."
	docker build -t learn-prod-service .

docker-run: docker-build
	@echo "Running in Linux container (privileged for pprof)..."
	docker run -it --rm -p 8080:8080 -p 9000:9000 learn-prod-service

clean:
	rm -f prod-service-patterns.db
	rm -f main
	rm -f server-linux
