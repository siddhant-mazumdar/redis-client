## Redis Client/Server (Go + HTTP + TCP)

A learning/utility project that exposes a Redis-like data store with:
- An HTTP API (Echo) for parsing RESP, storing/fetching data, and hash operations
- A Redis-compatible TCP server that understands a useful subset of Redis commands
- SQLite persistence with WAL for durability
- Clean/hexagonal architecture (use cases, adapters, ports)

This is not a drop-in replacement for Redis, but it’s handy for experimentation, demos, and local tooling.

---

### Features
- HTTP API under a configurable prefix (default: /apis/redis)
  - Parse RESP payloads and optionally persist results
  - Hash commands: HMSET, HMGET, multi-key HMGET
  - Command history endpoint
  - Health checks
- TCP server (default port 6379) with a subset of Redis commands:
  - PING, ECHO
  - SET, GET, MSET, MGET
  - DEL, EXISTS
  - INCR
  - Hashes: HSET, HGET, HEXISTS, HLEN, HGETALL, HKEYS, HVALS, HMSET, HMGET
  - Key search: KEYS, SCAN (MATCH/COUNT)
  - EXPIRE (basic TTL support) + periodic TTL cleanup
- Persistent storage using SQLite at ./data/redis.db (configurable)

---

### Repository layout
- cmd/               — HTTP and TCP server entrypoint
- internal/          — application core (use cases, ports/adapters)
  - infrastructure/  — input ports (http, tcp) and interface adapters (sqlite)
  - usecases/        — commands/queries and Redis TCP handlers
- config/            — Viper-based configuration (GO_ENV environments)
- main.go            — Example client to exercise the HTTP API (optional)
- data/              — SQLite DB files (created on first run)

---

### Prerequisites
- Go 1.21+ (module currently set to go 1.23)
- CGO enabled (github.com/mattn/go-sqlite3 requires a C toolchain)
- On Linux: ensure GCC is installed (e.g., build-essential)

---

### Quick start
1) Install dependencies
- go mod download

2) Start the server (HTTP + TCP)
- GO_ENV=development go run ./cmd
  - HTTP: http://localhost:10001
  - TCP:  127.0.0.1:6379

3) Health checks
- curl http://localhost:10001/health
- curl http://localhost:10001/health-check

4) Try the TCP server with redis-cli
- redis-cli -p 6379 PING
- redis-cli -p 6379 SET foo bar
- redis-cli -p 6379 GET foo
- redis-cli -p 6379 HMSET hkey1 f1 v1 f2 v2
- redis-cli -p 6379 HMGET hkey1 f1 f2
- redis-cli -p 6379 INCR counter
- redis-cli -p 6379 KEYS "*"
- redis-cli -p 6379 SCAN 0 MATCH "h*" COUNT 5

---

### Configuration
- Config is loaded from config/environments/${GO_ENV}/ (default GO_ENV=development)
  - common-config.json: core settings (e.g., sqliteDbPath)
  - config.json: service-specific values
  - service-endpoints.json: service endpoints and prefix
- Useful environment variables
  - PORT: HTTP port (default 10001)
  - REDIS_TCP_PORT: TCP port (default 6379)
  - SQLITE_MAX_CONNECTION, SQLITE_MAX_IDLE_CONNECTION: connection pool tuning

SQLite path (default): ./data/redis.db. The server will create the data directory if needed.

---

### HTTP API
Base URL: http://localhost:10001/apis/redis

- POST /v1/redis/parse-resp
  - Body: {"resp_data":"*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"}
  - Description: Parses RESP into key/value representation without persisting

- POST /v1/redis/parse-and-store
  - Body: {"resp_data":"*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"}
  - Description: Parses RESP and persists into SQLite

- POST /v1/redis/hmset
  - Body: {"key":"hkey1","fields":{"f1":"v1","f2":"v2"}}
  - Description: Set multiple hash fields

- POST /v1/redis/hmget
  - Body: {"key":"hkey1","fields":["f1","f2"]}
  - Description: Get specific hash fields

- POST /v1/redis/hmget-multi
  - Body: {"keys":{"hkey1":["f1","f2"],"hkey2":["x"]}}
  - Description: Multi-key hash fetch

- GET /v1/redis/command-history?limit=50
  - Description: Recent command history (server-side log)

Note: Additional key/value HTTP endpoints may evolve over time. Prefer the TCP interface for the broader Redis-like command set.

---

### Example cURL
- Parse RESP only
  - curl -s -X POST http://localhost:10001/apis/redis/v1/redis/parse-resp \
    -H "Content-Type: application/json" \
    -d '{"resp_data":"*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"}'

- Parse and store
  - curl -s -X POST http://localhost:10001/apis/redis/v1/redis/parse-and-store \
    -H "Content-Type: application/json" \
    -d '{"resp_data":"*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"}'

- Hashes
  - curl -s -X POST http://localhost:10001/apis/redis/v1/redis/hmset \
    -H "Content-Type: application/json" \
    -d '{"key":"hkey1","fields":{"f1":"v1","f2":"v2"}}'
  - curl -s -X POST http://localhost:10001/apis/redis/v1/redis/hmget \
    -H "Content-Type: application/json" \
    -d '{"key":"hkey1","fields":["f1","f2","f3"]}'
  - curl -s -X POST http://localhost:10001/apis/redis/v1/redis/hmget-multi \
    -H "Content-Type: application/json" \
    -d '{"keys":{"hkey1":["f1","f2"],"hkey2":["x"]}}'

---

### TCP examples

- Using redis-cli
  - redis-cli -p 6379 PING
  - redis-cli -p 6379 SET foo bar
  - redis-cli -p 6379 GET foo
  - redis-cli -p 6379 HMSET hkey1 f1 v1 f2 v2
  - redis-cli -p 6379 HMGET hkey1 f1 f2
  - redis-cli -p 6379 SCAN 0 MATCH "h*" COUNT 5

- Using netcat (inline protocol)
  - printf "PING\r\n" | nc -N 127.0.0.1 6379
  - printf "SET foo bar\r\n" | nc -N 127.0.0.1 6379
  - printf "GET foo\r\n" | nc -N 127.0.0.1 6379
  - Note: If your nc doesn't support -N, try -q 1 instead (nc -q 1 ...)

- Using netcat (RESP protocol)
  - printf "*1\r\n$4\r\nPING\r\n" | nc -N 127.0.0.1 6379
  - printf "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n" | nc -N 127.0.0.1 6379
  - printf "*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n" | nc -N 127.0.0.1 6379
  - printf "*6\r\n$5\r\nHMSET\r\n$5\r\nhkey1\r\n$2\r\nf1\r\n$2\r\nv1\r\n$2\r\nf2\r\n$2\r\nv2\r\n" | nc -N 127.0.0.1 6379
  - printf "*4\r\n$5\r\nHMGET\r\n$5\r\nhkey1\r\n$2\r\nf1\r\n$2\r\nf2\r\n" | nc -N 127.0.0.1 6379
  - printf "*6\r\n$4\r\nSCAN\r\n$1\r\n0\r\n$5\r\nMATCH\r\n$2\r\nh*\r\n$5\r\nCOUNT\r\n$1\r\n5\r\n" | nc -N 127.0.0.1 6379

- Minimal Go client (RESP PING)

```go
conn, _ := net.Dial("tcp", "127.0.0.1:6379")
defer conn.Close()
fmt.Fprintf(conn, "*1\r\n$4\r\nPING\r\n")
resp := make([]byte, 64)
n, _ := conn.Read(resp)
fmt.Println(string(resp[:n])) // +PONG\r\n
```

---

### Architecture notes
- Clean architecture style:
  - Use cases expose Commands and Queries (internal/usecases/redis/{commands,queries})
  - Interface adapters implement persistence (SQLite) and are injected via a bridge
  - Input ports: HTTP (Echo) and a Redis-compatible TCP server
- Periodic TTL cleanup runs every 30 minutes; EXPIRE is supported over TCP

---

### Development tips
- Start server: GO_ENV=development go run ./cmd
- Optional example client (runs some HTTP calls): go run .
  - Note: the example client is for experimentation and may not cover all current endpoints
- Build binary: go build -o bin/redis-server ./cmd

---

### Disclaimer
This project is for learning and internal tooling. It is not a production-ready Redis replacement.
