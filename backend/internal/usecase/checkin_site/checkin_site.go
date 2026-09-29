package checkin_site

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/checkin"
	"github.com/tortillaproduction/memory-tracker/internal/domain/site"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

// ErrSiteNotFound は、サイトが存在しない・他のユーザーのサイト・アーカイブ済みのいずれか。
// 他人のサイトの存在を推測されないよう、区別せずに返す。
var ErrSiteNotFound = errors.New("site not found")

// duplicateWindow はこの時間内の同じサイトへの再チェックインを記録しない。
// 通知リンクやゲートの「開く」の連打・リトライで記録が増え、サイト別ストリークが
// 水増しされるのを防ぐ。
const duplicateWindow = 5 * time.Minute

// Usecase は /go/:siteId が踏まれたときや、ゲートで「開く」が押されたときに呼ばれる。
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
// 直近duplicateWindow以内に同じサイトのチェックインがあれば記録せず、URLだけを返す。
func (uc *Usecase) Execute(ctx context.Context, userID user.ID, siteID site.ID) (redirectURL string, err error) {
	s, err := uc.siteRepo.FindByID(ctx, siteID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrSiteNotFound
	}
	if err != nil {
		return "", err
	}
	if s.UserID() != userID || s.IsArchived() {
		return "", ErrSiteNotFound
	}

	c := checkin.NewCheckIn(uc.idGenerator.NewCheckInID(), userID, siteID)
	if _, err := uc.checkinRepo.SaveUnlessRecent(ctx, c, duplicateWindow); err != nil {
		return "", err
	}

	return s.URL(), nil
}
