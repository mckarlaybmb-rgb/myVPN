# Phase 3 Telegram Automation

## Environment

Set `TELEGRAM_ENABLED=true`, `TELEGRAM_BOT_TOKEN`, and optionally `TELEGRAM_ADMIN_CHAT_ID`. The bot uses Telegram long polling. A user links an existing application account with `/start email@example.com`; account data is never selected by Telegram ID alone without this link.

Supported commands are `/start`, `/account`, `/vpn`, `/status`, `/renew`, and `/support`. Telegram identity and notification history are stored in PostgreSQL. The unique notification index makes repeated scheduler runs safe.

## Database

Migration `003_phase3.sql` adds timestamps, node health fields, Telegram tables, and indexes. Migrations are additive and safe to run repeatedly.

## Operations and risks

The bot process needs outbound HTTPS access to `api.telegram.org`. Telegram delivery failures are logged by the caller and do not stop API or scheduler processes. Protect the bot token and API key through the deployment secret store. Long polling should be replaced or fronted by a webhook when horizontally scaling bot instances.
