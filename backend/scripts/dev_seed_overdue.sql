-- 開発用シード: 指定ユーザーに「期限切れ」サイトと「期限内」サイトを作り、プッシュ通知をONにする。
-- 何度実行しても、再び通知対象になる(通知ログを消して作り直す)。
--
-- 使い方(リポジトリルートから。emailはGoogleログインに使ったアドレス):
--   docker compose exec -T db psql -U postgres memorytracker -v email=you@example.com \
--     < backend/scripts/dev_seed_overdue.sql
--
-- 注意: プッシュ購読(push_subscriptions)はブラウザ固有なので、ここでは作らない。
-- 先にブラウザでログインし、UIのトグルでPushを有効にしておくこと。
-- 後片付けは dev_seed_cleanup.sql。
\set ON_ERROR_STOP on

BEGIN;

-- 対象ユーザーが存在しない場合、\gset が "no rows returned" で失敗し、ON_ERROR_STOPにより何も変更せず終了する
SELECT id AS user_id FROM users WHERE email = :'email' \gset

-- 前回のシードを掃除(CASCADEでcheck_ins/notification_logsも消える)
DELETE FROM sites WHERE id LIKE 'dev-seed-%';

INSERT INTO sites (id, user_id, name, url, interval_hours) VALUES
  ('dev-seed-overdue-checked',   :'user_id', '[seed] 期限切れ(チェックイン済み)', 'https://example.com/overdue-checked',   1),
  ('dev-seed-overdue-unchecked', :'user_id', '[seed] 期限切れ(未チェックイン)',   'https://example.com/overdue-unchecked', 1),
  ('dev-seed-fresh',             :'user_id', '[seed] 期限内(通知されない)',        'https://example.com/fresh',             1);

INSERT INTO check_ins (id, user_id, site_id, checked_at, is_initial) VALUES
  ('dev-seed-ci-overdue-checked-initial',   :'user_id', 'dev-seed-overdue-checked',   now() - interval '5 hours', true),
  ('dev-seed-ci-overdue-checked',           :'user_id', 'dev-seed-overdue-checked',   now() - interval '3 hours', false),
  ('dev-seed-ci-overdue-unchecked-initial', :'user_id', 'dev-seed-overdue-unchecked', now() - interval '3 hours', true),
  ('dev-seed-ci-fresh-initial',             :'user_id', 'dev-seed-fresh',             now(),                      true);

-- プッシュ通知ON(有効な購読があればメールは止める既定動作)
INSERT INTO notification_settings (id, user_id, email_enabled, push_enabled, disable_email_when_push_available)
VALUES ('dev-seed-ns-' || :'user_id', :'user_id', true, true, true)
ON CONFLICT (user_id) DO UPDATE
  SET push_enabled = true, disable_email_when_push_available = true;

COMMIT;

SELECT s.name, s.interval_hours, max(c.checked_at) AS last_checkin
FROM sites s LEFT JOIN check_ins c ON c.site_id = s.id
WHERE s.id LIKE 'dev-seed-%' GROUP BY s.id ORDER BY s.id;
