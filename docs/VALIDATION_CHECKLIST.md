# Validation Checklist

## Implemented MVP Checks

- Run `git diff --check`.
- Run `go test ./...` from `backend`.
- Validate Compose with `docker-compose config`.
- Start PostgreSQL and the API with `docker-compose up --build`.
- Confirm `GET /health` returns HTTP 200.
- Create a user with `POST /api/v1/users` and confirm the row exists.
- Create a subscription with `POST /api/v1/subscriptions` and confirm its queue row starts as `pending`.
- Confirm a queue consumer can claim, complete, and fail jobs as expected.

## Planned Checks

- Run a persistent worker against a test Xray implementation once that worker exists.
- Validate retry, dead-letter, metrics, billing, and multi-node workflows after those features are implemented.
