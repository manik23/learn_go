FROM debian:bookworm-slim

# Install networking and tracing tools
RUN apt-get update && apt-get install -y \
    strace \
    lsof \
    net-tools \
    iproute2 \
    tcpdump \
    procps \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the pre-built linux binary
COPY server-linux /app/server

# Ensure it's executable
RUN chmod +x /app/server

# Default to tracing the network, read, and write syscalls
ENTRYPOINT ["strace", "-e", "trace=network,read,write", "-f"]
CMD ["./server"]
