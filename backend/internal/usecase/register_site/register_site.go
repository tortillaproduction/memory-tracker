package register_site

import (
	"context"
	"errors"
	"net/url"

	"github.com/tortillaproduction/memory-tracker/internal/domain/checkin"
	"github.com/tortillaproduction/memory-tracker/internal/domain/site"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

var ErrSiteLimitReached = errors.New("site limit reached for current plan")
var ErrInvalidURL = errors.New("url must be a valid absolute URL")
var ErrInvalidInterval = errors.New("interval hours must be between 1 and 8760")

// maxIntervalHours はチェック間隔の上限（1年分）。フロントのプリセット(最大168時間)より
// 十分大きく取りつつ、APIを直接叩かれた場合に桁違いの値(例: UNIXタイムスタンプの誤入力)が
// 保存されてしまうのを防ぐための現実的な上限。
const maxIntervalHours = 24 * 365

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
	parsed, err := url.ParseRequestURI(in.URL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, ErrInvalidURL
	}
	if in.IntervalHours <= 0 || in.IntervalHours > maxIntervalHours {
		return nil, ErrInvalidInterval
	}

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
	// 登録時のチェックインはis_initial=trueにし、ストリーク計算の対象外とする。
	initialCheckIn := checkin.NewInitialCheckIn(uc.idGenerator.NewCheckInID(), in.UserID, newSite.ID())
	if err := uc.checkinRepo.Save(ctx, initialCheckIn); err != nil {
		return nil, err
	}

	return newSite, nil
}
