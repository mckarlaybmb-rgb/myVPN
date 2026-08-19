# Database Schema - VPN OS v1

The canonical schema is [001_initial.sql](../backend/migrations/001_initial.sql). Migrations run in filename order when the API starts.

## Tables

### users

- `id UUID PRIMARY KEY`
- `email TEXT NOT NULL UNIQUE`
- `created_at TIMESTAMPTZ NOT NULL`

### subscriptions

- `id UUID PRIMARY KEY`
- `user_id UUID NOT NULL REFERENCES users(id)`
- `plan TEXT NOT NULL`
- `status TEXT NOT NULL DEFAULT 'active'`
- `expires_at TIMESTAMPTZ NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL`

### vpn_nodes

- `id UUID PRIMARY KEY`
- `name`, `address`, and `status` identify a managed node
- `port INTEGER NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL`

### job_queue

- `id UUID PRIMARY KEY`
- `job_type TEXT NOT NULL`
- `entity_id UUID NULL`
- `payload JSONB NOT NULL`
- `status` is constrained to `pending`, `processing`, `completed`, or `failed`
- `attempts`, `last_error`, `available_at`, and timestamps support durable processing

The queue has a partial index over pending jobs by availability and creation time. Future migrations should be versioned SQL files and should preserve existing data.
