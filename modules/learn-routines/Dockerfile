# ---------- Builder stage ----------
FROM golang:1.25-alpine AS builder

WORKDIR /src

# Install git (needed for go mod download in many cases)
RUN apk add --no-cache git

# Copy go mod files first (better cache usage)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY src .


# Build binary
# Adjust ./cmd/app if your main package path differs
RUN CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 \
    go build \
    -ldflags "\
    -s -w \
    -X 'main.Version=${VERSION}' \
    -X 'main.GitCommit=${GIT_COMMIT}' \
    -X 'main.BuildTime=${BUILD_TIME}'" \
    -o /out/learn-routines ./...

# ---------- Final stage (artifact only) ----------
FROM scratch AS export
COPY --from=builder /out/learn-routines /learn-routines
