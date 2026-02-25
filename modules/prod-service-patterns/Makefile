.PHONY: run bench profile-cpu profile-heap clean

# Parameters
PORT=8080
PPROF_PORT=9000
SERVER_URL=http://localhost:$(PORT)/process
PPROF_URL=http://localhost:$(PPROF_PORT)/debug/pprof

run:
	@echo "Starting production service on :$(PORT)..."
	@echo "pprof enabled on :$(PPROF_PORT)..."
	go run main.go

bench:
	@echo "Running local benchmarks (logic only)..."
	go test -v -bench=BenchmarkHandleProcessLocal -benchmem -run=^$$ . -benchtime=30s

bench-remote:
	@echo "Running remote benchmarks (network + logic)..."
	@echo "Make sure the server is running in another terminal!"
	go test -v -bench=BenchmarkHandleProcessRemote -benchmem -run=^$$ . -benchtime=30s

profile-cpu:
	@echo "Capturing 30s CPU profile..."
	go tool pprof -http=:8001 "$(PPROF_URL)/profile?seconds=30"

profile-heap:
	@echo "Capturing current heap profile..."
	go tool pprof -http=:8002 "$(PPROF_URL)/heap"

clean:
	rm -f prod-service-patterns.db
	rm -f main
