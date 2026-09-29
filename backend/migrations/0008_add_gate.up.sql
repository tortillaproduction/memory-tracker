-- ゲート(Androidアプリ)用の認証トークン。1ユーザーにつき1つで、再発行すると置き換わる。
-- トークンの平文は保存せず、SHA-256のハッシュだけを持つ。
-- last_used_at: アプリからAPIを最後に呼んだ時刻。一定期間更新がなければ端末未接続とみなし、メールを併用する。
CREATE TABLE gate_tokens (
    id            TEXT PRIMARY KEY,
    user_id       TEXT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    token_hash    TEXT NOT NULL UNIQUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at  TIMESTAMPTZ
);

-- ゲートの脱出口(サイトを開かずに解除)を使った記録。dismissed_onは日本時間の日付。
-- 1日1件で、同じ日に何度解除しても行は増えない。
CREATE TABLE gate_dismissals (
    id            TEXT PRIMARY KEY,
    user_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    dismissed_on  DATE NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, dismissed_on)
);

-- 通知モード。'gate'のユーザーには原則メールを送らない(ゲート端末が未接続の場合を除く)。
ALTER TABLE notification_settings
    ADD COLUMN mode TEXT NOT NULL DEFAULT 'email' CHECK (mode IN ('email', 'gate'));
