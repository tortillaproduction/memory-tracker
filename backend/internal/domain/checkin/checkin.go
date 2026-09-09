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
	isInitial bool // 登録時の初回チェックインはtrue。ストリーク計算から除外する。
}

func NewCheckIn(id ID, userID user.ID, siteID site.ID) *CheckIn {
	return &CheckIn{
		id:        id,
		userID:    userID,
		siteID:    siteID,
		checkedAt: time.Now(),
		isInitial: false,
	}
}

// NewInitialCheckIn は「登録=チェックイン」時に呼ぶ。ストリーク計算の対象外になる。
func NewInitialCheckIn(id ID, userID user.ID, siteID site.ID) *CheckIn {
	return &CheckIn{
		id:        id,
		userID:    userID,
		siteID:    siteID,
		checkedAt: time.Now(),
		isInitial: true,
	}
}

// Reconstruct はDBから読み出した値でCheckInを復元する。
// checkedAtをtime.Now()で上書きせず、保存済みの値をそのまま使う点がNewCheckInと異なる。
func Reconstruct(id ID, userID user.ID, siteID site.ID, checkedAt time.Time, isInitial bool) *CheckIn {
	return &CheckIn{
		id:        id,
		userID:    userID,
		siteID:    siteID,
		checkedAt: checkedAt,
		isInitial: isInitial,
	}
}

func (c *CheckIn) ID() ID               { return c.id }
func (c *CheckIn) UserID() user.ID      { return c.userID }
func (c *CheckIn) SiteID() site.ID      { return c.siteID }
func (c *CheckIn) CheckedAt() time.Time { return c.checkedAt }
func (c *CheckIn) IsInitial() bool      { return c.isInitial }

type Repository interface {
	Save(ctx context.Context, c *CheckIn) error
	FindLatestBySiteID(ctx context.Context, siteID site.ID) (*CheckIn, error)
	FindAllByUserID(ctx context.Context, userID user.ID, since time.Time) ([]*CheckIn, error)
	FindAllBySiteID(ctx context.Context, siteID site.ID) ([]*CheckIn, error)
}
