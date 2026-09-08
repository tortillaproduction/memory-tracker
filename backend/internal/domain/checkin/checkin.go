package checkin

import (
	"context"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/site"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

type ID string

// CheckIn は「いつ・どのユーザーが・どのサイトを訪問したか」を表すイベント的エンティティ。
// ストリーク計算は全てこのログから導出する（別テーブルにストリーク値を持たない）。
type CheckIn struct {
	id        ID
	userID    user.ID
	siteID    site.ID
	checkedAt time.Time
}

func NewCheckIn(id ID, userID user.ID, siteID site.ID) *CheckIn {
	return &CheckIn{
		id:        id,
		userID:    userID,
		siteID:    siteID,
		checkedAt: time.Now(),
	}
}

func (c *CheckIn) ID() ID               { return c.id }
func (c *CheckIn) SiteID() site.ID      { return c.siteID }
func (c *CheckIn) CheckedAt() time.Time { return c.checkedAt }

type Repository interface {
	Save(ctx context.Context, c *CheckIn) error
	FindLatestBySiteID(ctx context.Context, siteID site.ID) (*CheckIn, error)
	FindAllByUserID(ctx context.Context, userID user.ID, since time.Time) ([]*CheckIn, error)
}
