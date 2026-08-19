# API Specification

Base path: `/api/v1`

## Health

`GET /health` returns `{"status":"ok"}`.

## Users

`GET /api/v1/users` lists users ordered by creation time.

`POST /api/v1/users` creates a user.

Request:

```json
{"email":"user@example.com"}
```

The response is the created user with `id`, `email`, and `created_at`.

## Subscriptions

`GET /api/v1/subscriptions` lists subscriptions ordered by creation time.

`POST /api/v1/subscriptions` creates a subscription and enqueues a durable job.

Request:

```json
{"user_id":"<uuid>","plan":"monthly","expires_at":"2026-09-19T00:00:00Z"}
```

The response is the created subscription. Invalid JSON or missing required fields returns `400`; database failures return `500`; unsupported methods return `405`.
