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
	@echo "Building Control Plane Linux Binary locally..."
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o control-plane-linux main.go
	@echo "Building Control Plane Linux Container..."
	docker build -t learn-control-plane .

docker-run: docker-build
	@echo "Running Control Plane in Linux Environment..."
	docker run -it --rm -p 8080:8080 learn-control-plane

cluster-run: docker-build
	@echo "Starting 3-Node Sharded Control Plane Cluster..."
	@touch $(PWD)/test.db
	docker run -d --name cp-node-1 -v $(PWD)/test.db:/app/test.db -e NODE_ID=node-1 -e NODE_INDEX=0 -e TOTAL_NODES=3 -p 8080:8080 learn-control-plane
	docker run -d --name cp-node-2 -v $(PWD)/test.db:/app/test.db -e NODE_ID=node-2 -e NODE_INDEX=1 -e TOTAL_NODES=3 -p 8081:8080 learn-control-plane
	docker run -d --name cp-node-3 -v $(PWD)/test.db:/app/test.db -e NODE_ID=node-3 -e NODE_INDEX=2 -e TOTAL_NODES=3 -p 8082:8080 learn-control-plane
	@echo "Cluster running — nodes own: node-1=shard0, node-2=shard1, node-3=shard2"

cluster-stop:
	@echo "Cleaning up cluster nodes..."
	@docker stop cp-node-1 cp-node-2 cp-node-3 2>/dev/null || true
	@docker rm cp-node-1 cp-node-2 cp-node-3 2>/dev/null || true

# Kill the current leader to test failover
# After 15s, node-2 or node-3 should take over the lease
cluster-kill-leader:
	@echo "Killing cp-node-1 (current leader) to test failover..."
	@docker stop cp-node-1
	@echo "Leader killed! Watch node-2 or node-3 take over after 15s..."
	@echo "Run: make cluster-logs to observe the takeover"

cluster-status:
	@echo "=== Cluster Health ==="
	@docker ps --filter "name=cp-node" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

cluster-logs:
	@echo "Tailing logs from all running nodes (Ctrl+C to stop)..."
	@for c in $$(docker ps --filter "name=cp-node" --format "{{.Names}}"); do \
		docker logs --tail=5 -f $$c 2>&1 | sed "s/^/[$$c] /" & \
	done; wait

clean:
	rm -f control-plane
	rm -f test.db
