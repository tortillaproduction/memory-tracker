ALTER TABLE notification_settings
    DROP COLUMN IF EXISTS push_enabled,
    DROP COLUMN IF EXISTS disable_email_when_push_available;
