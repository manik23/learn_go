FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY control-plane-linux ./control-plane

EXPOSE 8080

CMD ["./control-plane"]
