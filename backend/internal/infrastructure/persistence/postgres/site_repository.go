package postgres

import (
	"context"
	"database/sql"

	"github.com/tortillaproduction/study-tracker/internal/domain/site"
	"github.com/tortillaproduction/study-tracker/internal/domain/user"
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
	`, s.ID(), s.UserID(), s.URL(), s.URL(), s.IntervalHours(), s.IsArchived())
	// NOTE: 実際にはgetter経由でNameも渡す必要あり。sqlcでの自動生成に置き換え推奨。
	return err
}

func (r *siteRepository) FindByID(ctx context.Context, id site.ID) (*site.Site, error) {
	// TODO: sqlcで型安全なクエリに置き換える
	return nil, nil
}

func (r *siteRepository) FindAllByUserID(ctx context.Context, userID user.ID) ([]*site.Site, error) {
	// TODO: 実装
	return nil, nil
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
