# VPN OS v1 Architecture

## Phase 1 MVP

The backend is a Go 1.24 HTTP service using PostgreSQL as its source of truth. It runs database migrations on startup, exposes the health and CRUD endpoints under `/api/v1`, protects API routes with a development API key, and enqueues subscription events in the durable `job_queue` table.

```text
HTTP client -> Go API -> services -> PostgreSQL
                              |
                              +-> job_queue (pending -> processing -> completed/failed)
```

### Components

- `cmd/api`: application startup, dependency wiring, and HTTP server.
- `internal/config`: environment-based configuration.
- `internal/database`: environment-based PostgreSQL pool and startup migrations.
- `internal/handlers`: HTTP parsing, validation, and JSON responses.
- `internal/services`: user and subscription application logic.
- `internal/jobs`: durable queue states, claiming, completion, failure, and enqueue operations. No worker is started yet.
- `internal/middleware`: development `X-API-Key` authentication for API routes.
- `internal/repositories`: PostgreSQL query implementations.
- `internal/xray`: an interface boundary only; no Xray client is wired.
- `migrations`: idempotent initial schema for users, subscriptions, VPN nodes, and jobs.

### Running

From the repository root, run `docker compose up --build`. The API is available on port `8080`; PostgreSQL is available on port `5432`.

The service reads `DATABASE_URL`, `BACKEND_PORT`, and `ADMIN_API_KEY`. Subscription creation requires `user_id`, `plan`, and an RFC3339 `expires_at` value, and creates a `pending` job for future processing.

Xray HandlerService, StatsService, persistent workers, Prometheus, payment, Telegram, dashboard, and multi-node orchestration are not implemented in this phase.
