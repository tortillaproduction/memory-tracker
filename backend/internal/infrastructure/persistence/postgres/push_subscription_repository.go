package postgres

import (
	"context"
	"database/sql"

	"github.com/tortillaproduction/memory-tracker/internal/domain/notification"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

type pushSubscriptionRepository struct {
	db *sql.DB
}

func NewPushSubscriptionRepository(db *sql.DB) notification.PushSubscriptionRepository {
	return &pushSubscriptionRepository{db: db}
}

func (r *pushSubscriptionRepository) Create(ctx context.Context, sub *notification.PushSubscription) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO push_subscriptions (id, user_id, endpoint, p256dh_key, auth_key, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (endpoint) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			p256dh_key = EXCLUDED.p256dh_key,
			auth_key = EXCLUDED.auth_key,
			user_agent = EXCLUDED.user_agent,
			last_used_at = now()
	`, sub.ID(), sub.UserID(), sub.Endpoint(), sub.P256dhKey(), sub.AuthKey(), sub.UserAgent())
	return err
}

func (r *pushSubscriptionRepository) DeleteByEndpoint(ctx context.Context, endpoint string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM push_subscriptions WHERE endpoint = $1`, endpoint)
	return err
}

func (r *pushSubscriptionRepository) DeleteByID(ctx context.Context, userID user.ID, id notification.SubscriptionID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM push_subscriptions WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *pushSubscriptionRepository) ListByUserID(ctx context.Context, userID user.ID) ([]*notification.PushSubscription, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, endpoint, p256dh_key, auth_key, user_agent
		FROM push_subscriptions
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*notification.PushSubscription
	for rows.Next() {
		var (
			id, uid, endpoint, p256dh, auth string
			userAgent                       sql.NullString
		)
		if err := rows.Scan(&id, &uid, &endpoint, &p256dh, &auth, &userAgent); err != nil {
			return nil, err
		}
		result = append(result, notification.NewPushSubscription(
			notification.SubscriptionID(id), user.ID(uid), endpoint, p256dh, auth, userAgent.String,
		))
	}

	return result, rows.Err()
}
