# VPN OS v1 (vpnos)

VPN OS v1 is a production-ready backend control plane for Xray-core VPN nodes. It manages VLESS + REALITY + XTLS Vision users dynamically via a durable PostgreSQL-backed job queue, background workers with retries and dead-letter handling, and a small REST API for control-plane operations.

See docs/ for architecture and deployment details.
