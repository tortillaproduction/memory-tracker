-- 0006・0007と同じ定義でテーブルと列を作り直す。
-- 削除した購読情報と設定値は戻らない(列は既定値、テーブルは空で復元される)。
CREATE TABLE IF NOT EXISTS push_subscriptions (
    id            TEXT PRIMARY KEY,
    user_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint      TEXT NOT NULL,
    p256dh_key    TEXT NOT NULL,
    auth_key      TEXT NOT NULL,
    user_agent    TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_push_subscriptions_endpoint ON push_subscriptions(endpoint);
CREATE INDEX IF NOT EXISTS idx_push_subscriptions_user_id ON push_subscriptions(user_id);

ALTER TABLE notification_settings
    ADD COLUMN IF NOT EXISTS push_enabled BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS disable_email_when_push_available BOOLEAN NOT NULL DEFAULT true;
