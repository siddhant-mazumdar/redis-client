# Multi-stage build with CGO (sqlite3)

# 1) Builder stage
FROM golang:1.23-bullseye AS builder

# Install build deps for CGO (gcc, etc.)
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Cache modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the server binary
# Note: we set CGO_ENABLED=1 due to github.com/mattn/go-sqlite3
ENV CGO_ENABLED=1
RUN go build -o /app/bin/redis-lite-server ./cmd

# 2) Runtime stage
FROM debian:bullseye-slim

# Install runtime libs for CGO binaries
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates libgcc1 && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /srv

# Create non-root user
RUN useradd -m -u 10001 appuser

# Copy binary and any required files
COPY --from=builder /app/bin/redis-lite-server /usr/local/bin/redis-lite-server

# App data dir
RUN mkdir -p /srv/data && chown -R appuser:appuser /srv

USER appuser

# Default env (can be overridden)
ENV PORT=10001 \
    REDIS_TCP_PORT=6379 \
    GO_ENV=production

EXPOSE 10001 6379

# Entrypoint
ENTRYPOINT ["/usr/local/bin/redis-lite-server"]

