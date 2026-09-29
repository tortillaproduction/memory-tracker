-- ブラウザ通知(Web Push)の廃止に伴い、購読情報のテーブルと通知設定の関連列を削除する。
-- コードからの参照は先に削除済み(アプリはこれらを読み書きしない)。
DROP TABLE IF EXISTS push_subscriptions;

ALTER TABLE notification_settings
    DROP COLUMN IF EXISTS push_enabled,
    DROP COLUMN IF EXISTS disable_email_when_push_available;
