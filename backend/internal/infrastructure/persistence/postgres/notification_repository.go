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
		INSERT INTO notification_settings (id, user_id, email_enabled, line_enabled, line_user_id, mode)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, s.ID(), s.UserID(), s.EmailEnabled(), s.LineEnabled(), s.LineUserID(), string(s.Mode()))
	return err
}

func (r *notificationRepository) FindByUserID(ctx context.Context, userID user.ID) (*notification.Setting, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, email_enabled, line_enabled, line_user_id, mode
		FROM notification_settings
		WHERE user_id = $1
	`, userID)

	var (
		id           string
		uid          string
		emailEnabled bool
		lineEnabled  bool
		lineUserID   sql.NullString
		mode         string
	)
	if err := row.Scan(&id, &uid, &emailEnabled, &lineEnabled, &lineUserID, &mode); err != nil {
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
	s.SetMode(notification.Mode(mode))
	return s, nil
}

func (r *notificationRepository) Update(ctx context.Context, s *notification.Setting) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE notification_settings
		SET email_enabled = $2, line_enabled = $3, line_user_id = $4, mode = $5
		WHERE user_id = $1
	`, s.UserID(), s.EmailEnabled(), s.LineEnabled(), s.LineUserID(), string(s.Mode()))
	return err
}
