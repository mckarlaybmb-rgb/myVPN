# Xray Integration - Phase 1 Boundary

Phase 1 defines the integration boundary in `backend/internal/xray`. The `Client` interface provides `AddUser` and `RemoveUser` operations for future Xray HandlerService integration; no live Xray client is wired into the API yet.

## Planned Flow

1. A user or subscription operation is committed to PostgreSQL.
2. A durable `job_queue` record is created for the downstream operation.
3. A worker claims the job and invokes the Xray client for the selected `vpn_node`.
4. The worker marks the job `completed` or `failed`.

## Configuration Direction

The existing environment template contains Xray-related settings such as `XRAY_BIN`, `XRAY_API_ADDR`, and the VLESS + REALITY URI parameters. These are reserved for the integration implementation and are not required by the current API startup path.

Production deployments should keep Xray administration local to the node or behind a mutually authenticated channel.
