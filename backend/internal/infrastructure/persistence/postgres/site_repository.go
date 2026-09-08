package postgres

import (
	"context"
	"database/sql"

	"github.com/tortillaproduction/memory-tracker/internal/domain/site"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

// siteRepository は domain/site.Repository インターフェースの実装。
// domain層はこの型の存在を知らない（依存性逆転）。
type siteRepository struct {
	db *sql.DB
}

func NewSiteRepository(db *sql.DB) site.Repository {
	return &siteRepository{db: db}
}

func (r *siteRepository) Save(ctx context.Context, s *site.Site) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sites (id, user_id, name, url, interval_hours, is_archived)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			url = EXCLUDED.url,
			interval_hours = EXCLUDED.interval_hours,
			is_archived = EXCLUDED.is_archived
	`, s.ID(), s.UserID(), s.Name(), s.URL(), s.URL(), s.IntervalHours(), s.IsArchived())
	return err
}

func (r *siteRepository) FindByID(ctx context.Context, id site.ID) (*site.Site, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, url, interval_hours, is_archived
		FROM sites WHERE id = $1
	`, id)
	return scanSite(row)
}

func (r *siteRepository) FindAllByUserID(ctx context.Context, userID user.ID) ([]*site.Site, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, name, url, interval_hours, is_archived
		FROM sites WHERE user_id = $1 AND is_archived = false
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []*site.Site
	for rows.Next() {
		s, err := scanSite(rows)
		if err != nil {
			return nil, err
		}
		sites = append(sites, s)
	}

	return sites, rows.Err()
}

// scannerはsql.Rowとsql.Rowsの両方に対応させるための最小インターフェース
type scanner interface {
	Scan(dest ...any) error
}

func scanSite(row scanner) (*site.Site, error) {
	var id, userID, name, url string
	var intervalHours int
	var isArchived bool

	if err := row.Scan(&id, &userID, &name, &url, &intervalHours, &isArchived); err != nil {
		return nil, err
	}

	s := site.NewSite(site.ID(id), user.ID(userID), name, url, intervalHours)
	if isArchived {
		s.Archive()
	}
	return s, nil
}

func (r *siteRepository) CountByUserID(ctx context.Context, userID user.ID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM sites WHERE user_id = $1 AND is_archived = false
	`, userID).Scan(&count)
	return count, err
}

func (r *siteRepository) Delete(ctx context.Context, id site.ID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sites WHERE id = $1`, id)
	return err
}
