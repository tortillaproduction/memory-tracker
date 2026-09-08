package checkin_site

import (
	"context"

	"github.com/tortillaproduction/memory-tracker/internal/domain/checkin"
	"github.com/tortillaproduction/memory-tracker/internal/domain/site"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

// Usecase は /go/:siteId が踏まれたときに呼ばれる。
// 「チェックインを記録してから、実サイトのURLを返す」のがこのユースケースの責務。
// 実際のHTTPリダイレクトはinterface/http/handler側で行う。
type Usecase struct {
	siteRepo    site.Repository
	checkinRepo checkin.Repository
	idGenerator IDGenerator
}

type IDGenerator interface {
	NewCheckInID() checkin.ID
}

func NewUsecase(siteRepo site.Repository, checkinRepo checkin.Repository, idGen IDGenerator) *Usecase {
	return &Usecase{siteRepo: siteRepo, checkinRepo: checkinRepo, idGenerator: idGen}
}

// Execute はチェックインを記録し、リダイレクト先URLを返す。
func (uc *Usecase) Execute(ctx context.Context, userID user.ID, siteID site.ID) (redirectURL string, err error) {
	s, err := uc.siteRepo.FindByID(ctx, siteID)
	if err != nil {
		return "", err
	}

	c := checkin.NewCheckIn(uc.idGenerator.NewCheckInID(), userID, siteID)
	if err := uc.checkinRepo.Save(ctx, c); err != nil {
		return "", err
	}

	return s.URL(), nil
}
