# API Specification

Status: Phase 1 MVP
Status: Phase 2

Base path: `/api/v1`

All `/api/v1` requests require the development `X-API-Key` header matching `ADMIN_API_KEY`. `/health` is public.

## Health

`GET /health` returns `{"status":"ok"}`.

## Users

`GET /api/v1/users` lists users ordered by creation time.

`POST /api/v1/users` creates a user.
`POST /api/v1/users` creates a user and provisions an enabled VLESS client in Xray. If Xray provisioning fails, the user creation is rolled back.

Request:

```json
{"email":"user@example.com"}
```

The response is the created user with `id`, `email`, and `created_at`.

`DELETE /api/v1/users/{id}` deletes a user and returns HTTP 204.
`DELETE /api/v1/users/{id}` removes the user's Xray client, deletes the user, and returns HTTP 204.

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

`POST /api/v1/subscriptions/{id}/suspend` marks a subscription suspended and disables the user's Xray client.

`POST /api/v1/subscriptions/{id}/expire` marks a subscription expired and disables the user's Xray client. Expiry scheduling is outside the HTTP API; a worker or billing process should call this operation when `expires_at` is reached.

## Planned

Pagination, richer error envelopes, production authentication, node endpoints, metrics, and Xray health checks are planned and are not currently exposed.
Pagination, richer error envelopes, production authentication, node endpoints, metrics, and Xray health checks are planned and are not currently exposed. Client UUIDs and runtime configuration are stored in PostgreSQL, but are not returned by the user API.
