package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/checkin"
	"github.com/tortillaproduction/memory-tracker/internal/domain/site"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

type checkinRepository struct {
	db *sql.DB
}

func NewCheckInRepository(db *sql.DB) checkin.Repository {
	return &checkinRepository{db: db}
}

func (r *checkinRepository) Save(ctx context.Context, c *checkin.CheckIn) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO check_ins (id, user_id, site_id, checked_at, is_initial)
		VALUES ($1, $2, $3, $4, $5)
	`, c.ID(), c.UserID(), c.SiteID(), c.CheckedAt(), c.IsInitial())
	return err
}

func (r *checkinRepository) FindLatestBySiteID(ctx context.Context, siteID site.ID) (*checkin.CheckIn, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, site_id, checked_at, is_initial
		FROM check_ins
		WHERE site_id = $1 AND is_initial = false
		ORDER BY checked_at DESC
		LIMIT 1
	`, siteID)

	c, err := scanCheckIn(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // チェックイン未実施は nil で表現
	}
	return c, err
}

func (r *checkinRepository) FindAllByUserID(ctx context.Context, userID user.ID, since time.Time) ([]*checkin.CheckIn, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, site_id, checked_at, is_initial
		FROM check_ins
		WHERE user_id = $1 AND checked_at >= $2 AND is_initial = false
		ORDER BY checked_at DESC
	`, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*checkin.CheckIn
	for rows.Next() {
		c, err := scanCheckIn(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, c)
	}

	return results, rows.Err()
}

func (r *checkinRepository) FindAllBySiteID(ctx context.Context, siteID site.ID) ([]*checkin.CheckIn, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, site_id, checked_at, is_initial
		FROM check_ins
		WHERE site_id = $1 AND is_initial = false
		ORDER BY checked_at DESC
	`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*checkin.CheckIn
	for rows.Next() {
		c, err := scanCheckIn(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, c)
	}

	return results, rows.Err()
}

type checkinScanner interface {
	Scan(dest ...any) error
}

func scanCheckIn(row checkinScanner) (*checkin.CheckIn, error) {
	var id, userID, siteID string
	var checkedAt time.Time
	var isInitial bool
	if err := row.Scan(&id, &userID, &siteID, &checkedAt, &isInitial); err != nil {
		return nil, err
	}

	return checkin.Reconstruct(
		checkin.ID(id),
		user.ID(userID),
		site.ID(siteID),
		checkedAt,
		isInitial,
	), nil
}
