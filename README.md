# VPN OS v1 (vpnos)

VPN OS v1 is a Phase 1 Go backend MVP for managing users and subscriptions backed by PostgreSQL. It exposes a small API, protects development API routes with an API key, and records subscription events in a durable PostgreSQL job queue.

The current implementation does not include an Xray client, background worker, retry or dead-letter processing, metrics, billing, bots, dashboards, or multi-node orchestration.

See [docs/](docs/) for architecture, API, deployment, and validation details.
