package postgres

import (
	"context"
	"database/sql"

	"github.com/tortillaproduction/memory-tracker/internal/domain/notification"
)

type notificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) notification.Repository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, s *notification.Setting) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO notification_settings (id, user_id, email_enabled, line_enabled, line_user_id)
		VALUES ($1, $2,$3,$4,$5)
	`, s.ID(), s.UserID(), s.EmailEnabled(), s.LineEnabled(), s.LineUserID())
	return err
}
