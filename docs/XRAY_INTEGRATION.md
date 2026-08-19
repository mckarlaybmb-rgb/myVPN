# Xray Integration - Phase 2

Status: Implemented

The backend manages one VLESS client per user. `internal/xray.Service` is the application boundary for client lifecycle operations:

- `CreateClient` generates a UUID, builds the VLESS configuration, persists it in `xray_clients`, and adds the account to Xray.
- `DeleteClient` removes the account from Xray and then deletes its PostgreSQL record.
- `EnableClient` and `DisableClient` add or remove the account from the configured inbound and update the persisted `enabled` flag.

## Runtime

The production adapter is `HandlerServiceRuntime`. It connects to Xray's local gRPC HandlerService and uses `AlterInbound` with VLESS `AddUserOperation` and `RemoveUserOperation`. The configured inbound must already exist and have the Xray API enabled.

Required environment variables:

```text
XRAY_API_ADDR=127.0.0.1:12789
XRAY_INBOUND_TAG=vless-reality-tcp
```

The runtime currently assumes a trusted local gRPC endpoint and uses insecure transport credentials. Deployments that expose the API beyond localhost must put it behind a protected network path or add mTLS before doing so.

## Persistence

`xray_clients` stores the user ID, email used as the Xray account key, generated UUID, protocol, JSON configuration, enabled state, and timestamps. A unique user constraint prevents duplicate clients. The UUID is version 4 and the protocol is constrained to `vless`.

If runtime provisioning fails after the database insert, the service removes the inserted record and returns an error. User creation also removes the user record when provisioning fails. Runtime deletion is performed before deleting the database record so an unsuccessful Xray operation does not silently lose the account identity.

## Subscription Lifecycle

Suspending or expiring a subscription calls `DisableClient` for its user. Renewal restores the subscription status to `active`, but does not currently re-enable a client automatically; an explicit enable operation belongs in the billing or subscription worker once that workflow is added.

The current HTTP endpoints are documented in `API_SPEC.md`. Expiration scheduling is intentionally separate from the API and should be driven by a durable worker or billing integration.
# Xray Integration - Phase 1 Boundary

Status: Planned integration

Phase 1 defines the integration boundary in `backend/internal/xray`. The `Client` interface provides `AddUser` and `RemoveUser` operations for future Xray HandlerService integration; no live Xray client is wired into the API yet.

## Planned Flow

1. A user or subscription operation is committed to PostgreSQL.
2. A durable `job_queue` record is created for the downstream operation.
3. A worker claims the job and invokes the Xray client for the selected `vpn_node`.
4. The worker marks the job `completed` or `failed`.

## Configuration Direction

The existing environment template contains Xray-related settings such as `XRAY_BIN`, `XRAY_API_ADDR`, and the VLESS + REALITY URI parameters. These are reserved for the integration implementation and are not required by the current API startup path.

Production deployments should keep Xray administration local to the node or behind a mutually authenticated channel.

No Xray HandlerService, StatsService, VLESS URI generator, or runtime node client is currently implemented.
