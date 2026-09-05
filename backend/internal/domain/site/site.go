package site

import (
	"time"

	"github.com/tortillaproduction/study-tracker/internal/domain/user"
)

type ID string

type Site struct {
	id            ID
	userID        user.ID
	name          string
	url           string
	intervalHours int // 通知間隔（デフォルト24時間、サイトごとに変更可）
	isArchived    bool
	createdAt     time.Time
}

const DefaultIntervalHours = 24

func NewSite(id ID, userID user.ID, name, url string, intervalHours int) *Site {
	if intervalHours <= 0 {
		intervalHours = DefaultIntervalHours
	}
	return &Site{
		id:            id,
		userID:        userID,
		name:          name,
		url:           url,
		intervalHours: intervalHours,
		isArchived:    false,
		createdAt:     time.Now(),
	}
}

func (s *Site) ID() ID             { return s.id }
func (s *Site) UserID() user.ID    { return s.userID }
func (s *Site) Name() string       { return s.name }
func (s *Site) URL() string        { return s.url }
func (s *Site) IntervalHours() int { return s.intervalHours }
func (s *Site) IsArchived() bool   { return s.isArchived }

func (s *Site) Archive() { s.isArchived = true }

// IsOverdue は最終チェックイン時刻から見て、通知間隔を過ぎているかを判定する。
func (s *Site) IsOverdue(lastCheckedAt time.Time, now time.Time) bool {
	deadline := lastCheckedAt.Add(time.Duration(s.intervalHours) * time.Hour)
	return now.After(deadline)
}
