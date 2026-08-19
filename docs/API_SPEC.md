# API Specification

Status: Phase 1 MVP

Base path: `/api/v1`

All `/api/v1` requests require the development `X-API-Key` header matching `ADMIN_API_KEY`. `/health` is public.

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

`DELETE /api/v1/users/{id}` deletes a user and returns HTTP 204.

## Subscriptions

`GET /api/v1/subscriptions/{user_id}` lists subscriptions for a user ordered by creation time.

`POST /api/v1/subscriptions` creates a subscription and enqueues a durable job.

Request:

```json
{"user_id":"<uuid>","plan":"monthly","expires_at":"2026-09-19T00:00:00Z"}
```

The response is the created subscription. Invalid JSON or missing required fields returns `400`; database failures return `500`; unsupported methods return `405`.

`POST /api/v1/subscriptions/{id}/renew` extends a subscription.

Request:

```json
{"extra_days":30}
```

`extra_days` must be greater than zero. The renewal also enqueues a durable job.

## Planned

Pagination, richer error envelopes, production authentication, node endpoints, metrics, and Xray health checks are planned and are not currently exposed.
