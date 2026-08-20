# Production Deployment

Run PostgreSQL with durable storage and set `DATABASE_URL`, a non-default `ADMIN_API_KEY`, and the X-UI credentials through a secret manager. Set `XUI_BASE_URL`, `XUI_USERNAME`, `XUI_PASSWORD`, and `XUI_INBOUND_ID` for runtime provisioning.

Apply migrations by starting the API or by running the database migration step during deployment. Configure `TELEGRAM_ENABLED`, `TELEGRAM_BOT_TOKEN`, and `TELEGRAM_ADMIN_CHAT_ID` only when Telegram outbound HTTPS is available.

Expose the API through TLS and restrict access to admin routes. The current API-key middleware is suitable for an internal control plane but does not provide users, roles, key rotation, rate limiting, or audit logs. Use a single scheduler instance until distributed locking is added. Back up PostgreSQL, monitor X-UI availability, queue failures, node health, and process restarts.
