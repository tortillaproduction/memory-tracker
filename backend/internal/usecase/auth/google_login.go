package auth

import (
	"context"
	"errors"

	domainUser "github.com/tortillaproduction/study-tracker/internal/domain/user"
)

// GoogleUserInfo はinfrastructure/authパッケージへの直接依存を避けるための入力DTO。
// usecase層はGoogle固有の実装を知らず、必要な情報だけを受け取る。
type GoogleUserInfo struct {
	GoogleID string
	Email    string
	Name     string
}

type IDGenerator interface {
	NewUserID() domainUser.ID
}

type Usecase struct {
	userRepo    domainUser.Repository
	idGenerator IDGenerator
}

func NewUsecase(userRepo domainUser.Repository, idGen IDGenerator) *Usecase {
	return &Usecase{userRepo: userRepo, idGenerator: idGen}
}

// Execute はGoogleIDで既存ユーザーを探し、いなければ新規作成する（＝サインアップと兼ねる）。
func (uc *Usecase) Execute(ctx context.Context, info GoogleUserInfo) (*domainUser.User, error) {
	existing, err := uc.userRepo.FindByGoogleID(ctx, info.GoogleID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, domainUser.ErrNotFound) {
		return nil, err
	}

	newUser := domainUser.NewUser(uc.idGenerator.NewUserID(), info.GoogleID, info.Email, info.Name)
	if err := uc.userRepo.Save(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}
