# System Design - VPN OS v1

## Scope

The Phase 1 backend is a Go 1.24 HTTP service with PostgreSQL persistence. The API handlers validate requests and delegate database work to application services. Subscription creation also records a durable job for later Xray provisioning.

## Module Responsibilities

- `cmd/api`: starts the service, connects to PostgreSQL, runs migrations, and registers routes.
- `internal/handlers`: HTTP methods, JSON decoding, validation, and responses.
- `internal/services`: user and subscription business operations.
- `internal/database`: pgx connection pool and ordered SQL migration execution.
- `internal/jobs`: durable enqueue, claim, complete, and fail operations.
- `internal/xray`: interface boundary for future Xray node operations.

## Concurrency and Failure Handling

Jobs are claimed with PostgreSQL `FOR UPDATE SKIP LOCKED`, allowing multiple future workers to consume pending jobs safely. Every job records attempts and the last error. Phase 1 exposes the queue operations but does not start a background worker or implement retry scheduling.

## Security

Local Compose credentials are development defaults only. Production deployments should inject secrets at runtime, restrict database access, and add authenticated administrative API access with TLS.

## Testing Strategy

Run `go test ./...` from `backend`. Database integration tests should use an isolated PostgreSQL instance or Testcontainers when the test suite is added.
