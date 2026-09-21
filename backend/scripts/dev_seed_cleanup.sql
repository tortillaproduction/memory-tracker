-- dev_seed_overdue.sql が作ったサイトを削除する(check_ins/notification_logsはCASCADEで消える)。
-- notification_settings はユーザー本来の設定なので触らない。
--   docker compose exec -T db psql -U postgres memorytracker < backend/scripts/dev_seed_cleanup.sql
DELETE FROM sites WHERE id LIKE 'dev-seed-%';
