CREATE TABLE users (
    id          TEXT PRIMARY KEY,
    google_id   TEXT UNIQUE NOT NULL,
    email       TEXT UNIQUE NOT NULL,
    name        TEXT NOT NULL,
    plan_type   TEXT NOT NULL DEFAULT 'free',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sites (
    id              TEXT PRIMARY KEY,
    user_id         TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    url             TEXT NOT NULL,
    interval_hours  INTEGER NOT NULL DEFAULT 24,
    is_archived     BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sites_user_id ON sites(user_id) WHERE is_archived = false;

CREATE TABLE check_ins (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    site_id     TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    checked_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_checkins_site_id ON check_ins(site_id, checked_at DESC);
CREATE INDEX idx_checkins_user_id ON check_ins(user_id, checked_at DESC);

CREATE TABLE notification_settings (
    id              TEXT PRIMARY KEY,
    user_id         TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email_enabled   BOOLEAN NOT NULL DEFAULT true,
    -- LINE通知は未実装のため初期値false固定。実装完了後にUIのトグルを解放する。
    line_enabled    BOOLEAN NOT NULL DEFAULT false,
    line_user_id    TEXT
);

-- Cookieベースセッション用（golang.org/x/crypto等でハッシュ化したセッションIDを保持）
CREATE TABLE sessions (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Phase 2（Stripe連携）用。MVP時点ではテーブルのみ作成し、未使用でも構わない。
CREATE TABLE subscriptions (
    id                      TEXT PRIMARY KEY,
    user_id                 TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_type               TEXT NOT NULL,
    status                  TEXT NOT NULL,
    stripe_subscription_id  TEXT,
    current_period_end      TIMESTAMPTZ
);
