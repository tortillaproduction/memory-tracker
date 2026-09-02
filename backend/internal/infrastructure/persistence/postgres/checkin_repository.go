package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/tortillaproduction/study-tracker/internal/domain/checkin"
	"github.com/tortillaproduction/study-tracker/internal/domain/site"
	"github.com/tortillaproduction/study-tracker/internal/domain/user"
)

type checkinRepository struct {
	db *sql.DB
}

func NewCheckInRepository(db *sql.DB) checkin.Repository {
	return &checkinRepository{db: db}
}

func (r *checkinRepository) Save(ctx context.Context, c *checkin.CheckIn) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO check_ins (id, user_id, site_id, checked_at)
		VALUES ($1, $2, $3, $4)
	`, c.ID(), c.SiteID(), c.CheckedAt())
	return err
}

func (r *checkinRepository) FindLatestBySiteID(ctx context.Context, siteID site.ID) (*checkin.CheckIn, error) {
	// TODO: sqlcで型安全なクエリに置き換える
	return nil, nil
}

func (r *checkinRepository) FindAllByUserID(ctx context.Context, userID user.ID, since time.Time) ([]*checkin.CheckIn, error) {
	// TODO: 実装（ストリーク計算に使う）
	return nil, nil
}
