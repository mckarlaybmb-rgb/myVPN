# Integration Tests

Integration tests that require PostgreSQL belong in this package. `database_test.go` runs the migration smoke test when `DATABASE_URL` is set and skips cleanly otherwise. Unit tests remain self-contained so `go test ./...` is runnable without external services.