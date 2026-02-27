FROM golang:1.24-bookworm AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o server main.go

FROM debian:bookworm-slim

# Install sqlite and profiling tools
RUN apt-get update && apt-get install -y \
    sqlite3 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/server .
# Copy DB if it exists, but usually it's created on start
# COPY prod-service-patterns.db . 

EXPOSE 8080 9000

CMD ["./server"]
