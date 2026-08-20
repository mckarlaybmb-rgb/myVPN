ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE vpn_nodes ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE vpn_nodes ADD COLUMN IF NOT EXISTS last_check_at TIMESTAMPTZ;
ALTER TABLE vpn_nodes ADD COLUMN IF NOT EXISTS latency_ms INTEGER;

CREATE TABLE IF NOT EXISTS telegram_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), telegram_id BIGINT NOT NULL UNIQUE,
    username TEXT NOT NULL DEFAULT '', first_name TEXT NOT NULL DEFAULT '', last_name TEXT NOT NULL DEFAULT '',
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS telegram_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), telegram_id BIGINT NOT NULL,
    notification_type TEXT NOT NULL, reference_id UUID, sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), status TEXT NOT NULL DEFAULT 'sent'
);
CREATE UNIQUE INDEX IF NOT EXISTS telegram_notifications_once_idx ON telegram_notifications (telegram_id, notification_type, reference_id);
CREATE UNIQUE INDEX IF NOT EXISTS telegram_notifications_once_nullable_idx ON telegram_notifications (telegram_id, notification_type, COALESCE(reference_id, '00000000-0000-0000-0000-000000000000'::uuid));
CREATE INDEX IF NOT EXISTS subscriptions_user_id_idx ON subscriptions (user_id);
CREATE INDEX IF NOT EXISTS subscriptions_status_idx ON subscriptions (status);
CREATE INDEX IF NOT EXISTS subscriptions_expires_at_idx ON subscriptions (expires_at);
CREATE INDEX IF NOT EXISTS telegram_users_user_id_idx ON telegram_users (user_id);
CREATE INDEX IF NOT EXISTS vpn_nodes_status_idx ON vpn_nodes (status);