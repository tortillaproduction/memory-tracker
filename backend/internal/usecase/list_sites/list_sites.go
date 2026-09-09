package list_sites

import (
	"context"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/checkin"
	"github.com/tortillaproduction/memory-tracker/internal/domain/site"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

// SiteWithStatus はAPIレスポンス用のDTO。
// ドメインオブジェクトに表示用の計算済み値を付加したもの。
type SiteWithStatus struct {
	ID                  site.ID
	Name                string
	URL                 string
	IntervalHours       int
	LastCheckedAt       *time.Time // 一度もチェックインしてない場合nil
	HoursSinceLastCheck float64
	IsOverdue           bool
	SiteStreak          int
}

type Usecase struct {
	siteRepo    site.Repository
	checkinRepo checkin.Repository
}

func NewUsecase(siteRepo site.Repository, checkinRepo checkin.Repository) *Usecase {
	return &Usecase{siteRepo: siteRepo, checkinRepo: checkinRepo}
}

func (uc *Usecase) Execute(ctx context.Context, userID user.ID) ([]*SiteWithStatus, int, error) {
	sites, err := uc.siteRepo.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	now := time.Now()

	// ユーザー全体のストリーク計算用に過去90日分のチェックインを取得
	since := now.AddDate(0, 0, -90)
	allCheckIns, err := uc.checkinRepo.FindAllByUserID(ctx, userID, since)
	if err != nil {
		return nil, 0, err
	}
	userStreak := checkin.CalculateUserStreak(allCheckIns, now)

	results := make([]*SiteWithStatus, 0, len(sites))
	for _, s := range sites {
		latest, err := uc.checkinRepo.FindLatestBySiteID(ctx, s.ID())
		if err != nil {
			return nil, 0, err
		}

		status := &SiteWithStatus{
			ID:            s.ID(),
			Name:          s.Name(),
			URL:           s.URL(),
			IntervalHours: s.IntervalHours(),
		}

		if latest != nil {
			t := latest.CheckedAt()
			status.LastCheckedAt = &t
			status.HoursSinceLastCheck = now.Sub(t).Hours()
			status.IsOverdue = s.IsOverdue(t, now)

			// サイト別ストリーク計算
			siteCheckIns, err := uc.checkinRepo.FindAllBySiteID(ctx, s.ID())
			if err != nil {
				return nil, 0, err
			}
			status.SiteStreak = checkin.CalculateSiteStreak(siteCheckIns, s.IntervalHours(), now)
		} else {
			// 一度もチェックインしていない（登録直後）場合
			status.IsOverdue = false
			status.HoursSinceLastCheck = 0
			status.SiteStreak = 0
		}

		results = append(results, status)
	}

	return results, userStreak, nil
}
