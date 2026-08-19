# Subscription System

Status: Phase 1 MVP

## Current Model

A subscription belongs to a user and contains a plan label, status, expiry timestamp, and creation timestamp. Phase 1 provides list and create endpoints and enqueues a `subscription.created` job after a successful insert.

## Create Flow

1. Validate `user_id`, `plan`, and RFC3339 `expires_at`.
2. Insert the subscription in PostgreSQL.
3. Enqueue a pending job containing the subscription and user identifiers.
4. Return the created subscription to the caller.

The database foreign key prevents subscriptions for missing users. Renewal is implemented through `POST /api/v1/subscriptions/{id}/renew` and enqueues a `subscription.renewed` job.

## Planned

Billing, cancellation, plan catalogs, expiry workers, notifications, and bandwidth enforcement are planned work rather than current behavior.
