package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/tortillaproduction/memory-tracker/internal/domain/notification"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

type notificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) notification.SettingRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, s *notification.Setting) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO notification_settings (id, user_id, email_enabled, line_enabled, line_user_id, push_enabled, disable_email_when_push_available)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, s.ID(), s.UserID(), s.EmailEnabled(), s.LineEnabled(), s.LineUserID(), s.PushEnabled(), s.DisableEmailWhenPushAvailable())
	return err
}

func (r *notificationRepository) FindByUserID(ctx context.Context, userID user.ID) (*notification.Setting, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, email_enabled, line_enabled, line_user_id, push_enabled, disable_email_when_push_available
		FROM notification_settings
		WHERE user_id = $1
	`, userID)

	var (
		id                            string
		uid                           string
		emailEnabled                  bool
		lineEnabled                   bool
		lineUserID                    sql.NullString
		pushEnabled                   bool
		disableEmailWhenPushAvailable bool
	)
	if err := row.Scan(&id, &uid, &emailEnabled, &lineEnabled, &lineUserID, &pushEnabled, &disableEmailWhenPushAvailable); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notification.ErrSettingNotFound
		}
		return nil, err
	}

	s := notification.NewSetting(notification.ID(id), user.ID(uid))
	if !emailEnabled {
		s.SetEmailEnabled(false)
	}
	if lineEnabled {
		s.SetLineEnabled(true, lineUserID.String)
	}
	s.SetPushEnabled(pushEnabled)
	s.SetDisableEmailWhenPushAvailable(disableEmailWhenPushAvailable)

	return s, nil
}

func (r *notificationRepository) Update(ctx context.Context, s *notification.Setting) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE notification_settings
		SET email_enabled = $2, line_enabled = $3, line_user_id = $4, push_enabled = $5, disable_email_when_push_available = $6
		WHERE user_id = $1
	`, s.UserID(), s.EmailEnabled(), s.LineEnabled(), s.LineUserID(), s.PushEnabled(), s.DisableEmailWhenPushAvailable())
	return err
}
