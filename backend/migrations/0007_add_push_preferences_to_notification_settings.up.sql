-- push_enabled: このユーザーが有効なプッシュ購読を持っているか(購読/解除のたびに更新)。
-- disable_email_when_push_available: プッシュが届く間はメール通知を止めるかどうか。
-- デフォルトtrueにすることで、PWAインストール後にメールとプッシュが二重に届く煩わしさを
-- 標準で避ける(既存ユーザーはpush_enabled=falseのままなので挙動は変わらない)。
ALTER TABLE notification_settings
    ADD COLUMN push_enabled BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN disable_email_when_push_available BOOLEAN NOT NULL DEFAULT true;
