# Appointment Service Boilerplate

Scaffold for building appointment-driven services with Go, PostgreSQL, and Chi. Provides production-friendly defaults for API key auth, rate limiting, observability, and graceful ops.

## Features

- REST API for appointments with filtering and validation
- API key authentication plus per-key rate limiting
- Structured JSON logging with request IDs
- Health/readiness endpoints and graceful shutdown
- Parameterized queries against PostgreSQL
- Example Podman Compose setup for local/production use

## Quick Start

```bash
# Install dependencies
go mod download

# Copy env template and fill in values
cp .env.example .env

# Run the server
go run ./cmd/server
```

### Podman Compose

```bash
podman compose up --build -d
podman compose logs -f
curl http://localhost:8080/health
podman compose down
```

## API Overview

### Health
- `GET /health` – liveness
- `GET /ready` – readiness (checks database)

### Appointments (requires `X-API-Key`)
- `GET /appointments?date=YYYY-MM-DD&customer_id=&provider_id=&branch=`
- `GET /appointments/{id}`
- `POST /appointments`

### Sample Request

```bash
curl -H "X-API-Key: demo" \
  "http://localhost:8080/appointments?date=2024-05-01"
```

## API Docs

- Install the generator once: `go install github.com/swaggo/swag/cmd/swag@latest`
- Regenerate docs after handler/model changes: `$(go env GOPATH)/bin/swag init -g cmd/server/main.go -o docs`
- Browse Swagger UI at `http://localhost:8080/swagger/index.html`

## Configuration

Required variables in `.env`:
- `PG_HOST`, `PG_PORT`, `PG_USER`, `PG_PASSWORD`, `PG_DB`
- `PG_HOST_DOCKER` (host name for the app container to reach Postgres, defaults to `postgres`)
- `API_KEYS` (comma-separated `service:key` pairs)

Optional:
- `SERVER_PORT` (default `8080`)
- `LOG_LEVEL` (`debug|info|warn|error`)
- `PG_SSLMODE`

Generate API keys:
```bash
go run ./cmd/keygen -service api
```

## Project Layout

```
cmd/
  server/   - HTTP entrypoint
  keygen/   - helper for API keys
internal/
  config/       - env loading
  db/           - Postgres client
  handlers/     - HTTP handlers
  middlewares/  - auth, rate limit, request ID
  models/       - request/response structs
  routes/       - chi router wiring
  services/     - business logic
  httputil/     - response helpers
  logger/       - slog setup
```

## Operations

- `go test ./...` to run tests
- Adjust rate limit in `cmd/server/main.go`
- Update Podman image tag and metadata in `Dockerfile`

## Resetting Git

To initialize a fresh repository:
```bash
rm -rf .git
git init -b main
git add .
git commit -m "Initial commit"
```
