package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/gate"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

type gateDismissalRepository struct {
	db *sql.DB
}

func NewGateDismissalRepository(db *sql.DB) gate.DismissalRepository {
	return &gateDismissalRepository{db: db}
}

// day は日付部分だけを使う(dismissed_onはDATE)。呼び出し側で日本時間の日に揃えておくこと。
func (r *gateDismissalRepository) Record(ctx context.Context, id gate.DismissalID, userID user.ID, day time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO gate_dismissals (id, user_id, dismissed_on)
		VALUES ($1, $2, $3::date)
		ON CONFLICT (user_id, dismissed_on) DO NOTHING
	`, id, userID, day.Format("2006-01-02"))
	return err
}

func (r *gateDismissalRepository) ExistsOn(ctx context.Context, userID user.ID, day time.Time) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (SELECT 1 FROM gate_dismissals WHERE user_id = $1 AND dismissed_on = $2::date)
	`, userID, day.Format("2006-01-02")).Scan(&exists)
	return exists, err
}
