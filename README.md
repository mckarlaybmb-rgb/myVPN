# VPN OS v1 (vpnos)

VPN OS v1 is a Phase 1 Go backend MVP for managing users and subscriptions backed by PostgreSQL. It exposes a small API, protects development API routes with an API key, and records subscription events in a durable PostgreSQL job queue.

The backend includes X-UI VLESS client lifecycle management, a retrying PostgreSQL worker, hourly subscription expiry automation, five-minute node health checks, Telegram long polling, and API-key-protected admin read endpoints.

See [docs/](docs/) for architecture, API, deployment, validation, Telegram, and background job details.
