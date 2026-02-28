# OS Detection
OS := $(shell uname)

.PHONY: run watch-state test docker-build docker-run clean

run:
	@echo "Starting Control Plane Node ($(OS))..."
	go run main.go

watch-state:
	@echo "Watching Desired vs Observed State..."
	@watch -n 1 "curl -s localhost:8080/state | python3 -m json.tool || true"

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
