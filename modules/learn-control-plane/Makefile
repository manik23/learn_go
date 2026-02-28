# OS Detection
OS := $(shell uname)

.PHONY: run watch-state test docker-build docker-run clean

run:
	@echo "Starting Control Plane Node ($(OS))..."
	go run main.go

watch-state:
	@echo "Watching Desired vs Observed State... (Ctrl+C to stop)"
	@watch -n 1 "curl -s -H 'X-Auth-Token: secret' localhost:8080/v1/state | python3 -m json.tool || true"

scale-up:
	@echo "Scaling Desired State to 10..."
	@curl -X POST http://localhost:8080/v1/desired \
		-H "Content-Type: application/json" \
		-H "X-Auth-Token: secret" \
		-d '{"count": 10}'

scale-down:
	@echo "Scaling Desired State to 2..."
	@curl -X POST http://localhost:8080/v1/desired \
		-H "Content-Type: application/json" \
		-H "X-Auth-Token: secret" \
		-d '{"count": 2}'

test:
	@echo "Running local unit tests..."
	go test -v -short ./...

test-remote:
	@echo "Running remote integration tests (server must be active)..."
	go test -v -run=TestRemoteProvisioning ./rest/v1/...

docker-build:
	@echo "Building Control Plane Linux Container..."
	docker build -t learn-control-plane .

docker-run: docker-build
	@echo "Running Control Plane in Linux Environment..."
	docker run -it --rm -p 8080:8080 learn-control-plane

clean:
	rm -f control-plane
	rm -rf db/
