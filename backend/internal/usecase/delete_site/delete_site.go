package delete_site

import (
	"context"
	"errors"

	"github.com/tortillaproduction/memory-tracker/internal/domain/site"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

var (
	ErrNotFound  = errors.New("site not found")
	ErrForbidden = errors.New("not allowed to delete this site")
)

type Usecase struct {
	siteRepo site.Repository
}

func NewUsecase(siteRepo site.Repository) *Usecase {
	return &Usecase{siteRepo: siteRepo}
}

func (uc *Usecase) Execute(ctx context.Context, userID user.ID, siteID site.ID) error {
	s, err := uc.siteRepo.FindByID(ctx, siteID)
	if err != nil {
		return err
	}
	if s == nil {
		return ErrNotFound
	}
	// 他ユーザーのサイトを削除できないよう所有者チェック
	if s.UserID() != userID {
		return ErrForbidden
	}
	return uc.siteRepo.Delete(ctx, siteID)
}
