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
