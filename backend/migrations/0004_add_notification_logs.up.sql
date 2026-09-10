-- 通知済み記録テーブル。同じサイトへの二重送信を防ぐために使う。
-- バッチは「最後の通知からinterval_hours以上経過しているか」をここで確認する。
CREATE TABLE notification_logs (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    site_id     TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    sent_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notification_logs_site_id ON notification_logs(site_id, sent_at DESC);
