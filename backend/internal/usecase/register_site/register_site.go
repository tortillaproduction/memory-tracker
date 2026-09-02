package register_site

import (
	"context"
	"errors"

	"github.com/tortillaproduction/study-tracker/internal/domain/checkin"
	"github.com/tortillaproduction/study-tracker/internal/domain/site"
	"github.com/tortillaproduction/study-tracker/internal/domain/user"
)

var ErrSiteLimitReached = errors.New("site limit reached for current plan")

type Input struct {
	UserID        user.ID
	Name          string
	URL           string
	IntervalHours int
}

type Usecase struct {
	userRepo    user.Repository
	siteRepo    site.Repository
	checkinRepo checkin.Repository
	idGenerator IDGenerator
}

// IDGenerator はID生成の詳細（UUID等）をusecaseから隠蔽するための小さな抽象。
type IDGenerator interface {
	NewSiteID() site.ID
	NewCheckInID() checkin.ID
}

func NewUsecase(userRepo user.Repository, siteRepo site.Repository, checkinRepo checkin.Repository, idGen IDGenerator) *Usecase {
	return &Usecase{userRepo: userRepo, siteRepo: siteRepo, checkinRepo: checkinRepo, idGenerator: idGen}
}

func (uc *Usecase) Execute(ctx context.Context, in Input) (*site.Site, error) {
	u, err := uc.userRepo.FindByID(ctx, in.UserID)
	if err != nil {
		return nil, err
	}

	currentCount, err := uc.siteRepo.CountByUserID(ctx, in.UserID)
	if err != nil {
		return nil, err
	}

	// ビジネスルール（件数制限）はdomain層のPlanオブジェクトが知っている。
	// usecaseは「聞くだけ」で、数字そのものは知らない。
	if !u.Plan().CanRegisterMoreSites(currentCount) {
		return nil, ErrSiteLimitReached
	}

	newSite := site.NewSite(uc.idGenerator.NewSiteID(), in.UserID, in.Name, in.URL, in.IntervalHours)
	if err := uc.siteRepo.Save(ctx, newSite); err != nil {
		return nil, err
	}

	// 「登録=チェックイン」のルールをここで満たす。
	initialCheckIn := checkin.NewCheckIn(uc.idGenerator.NewCheckInID(), in.UserID, newSite.ID())
	if err := uc.checkinRepo.Save(ctx, initialCheckIn); err != nil {
		return nil, err
	}

	return newSite, nil
}
