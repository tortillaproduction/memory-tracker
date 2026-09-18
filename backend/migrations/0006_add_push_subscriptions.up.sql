-- ブラウザのWeb Push購読情報を保持するテーブル。1ユーザーが複数デバイス/ブラウザで
-- 購読すると複数行になる(endpointがデバイス/ブラウザごとに異なるため)。
CREATE TABLE push_subscriptions (
    id            TEXT PRIMARY KEY,
    user_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint      TEXT NOT NULL,
    p256dh_key    TEXT NOT NULL,
    auth_key      TEXT NOT NULL,
    user_agent    TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_push_subscriptions_endpoint ON push_subscriptions(endpoint);
CREATE INDEX idx_push_subscriptions_user_id ON push_subscriptions(user_id);
