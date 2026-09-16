package auth

import (
	"context"
	"errors"

	domainNotification "github.com/tortillaproduction/memory-tracker/internal/domain/notification"
	domainUser "github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

// GoogleUserInfo はinfrastructure/authパッケージへの直接依存を避けるための入力DTO。
// usecase層はGoogle固有の実装を知らず、必要な情報だけを受け取る。
type GoogleUserInfo struct {
	GoogleID string
	Email    string
	Name     string
	Picture  string
}

type IDGenerator interface {
	NewUserID() domainUser.ID
	NewNotificationSettingID() domainNotification.ID
}

type Usecase struct {
	userRepo         domainUser.Repository
	notificationRepo domainNotification.Repository
	idGenerator      IDGenerator
}

func NewUsecase(userRepo domainUser.Repository, notificationRepo domainNotification.Repository, idGen IDGenerator) *Usecase {
	return &Usecase{userRepo: userRepo, notificationRepo: notificationRepo, idGenerator: idGen}
}

// Execute はGoogleIDで既存ユーザーを探し、いなければ新規作成する（＝サインアップと兼ねる）。
func (uc *Usecase) Execute(ctx context.Context, info GoogleUserInfo) (*domainUser.User, error) {
	existing, err := uc.userRepo.FindByGoogleID(ctx, info.GoogleID)
	if err == nil {
		if existing.PictureURL() != info.Picture {
			existing.UpdatePicture(info.Picture)
			saved, _, err := uc.userRepo.Save(ctx, existing)
			if err != nil {
				return nil, err
			}
			return saved, nil
		}
		return existing, nil
	}
	if !errors.Is(err, domainUser.ErrNotFound) {
		return nil, err
	}

	// FindByGoogleIDとSaveの間はロックしていないため、同一Googleアカウントに対する
	// 同時ログインリクエストが両方とも「未存在」と判定してここに来ることがある
	// (TOCTOU競合)。SaveはgoogleIDをキーにUPSERTするため、後から来た方は
	// 相手が作成した行を返してもらう形になり、insertedはfalseになる。
	newUser := domainUser.NewUser(uc.idGenerator.NewUserID(), info.GoogleID, info.Email, info.Name, info.Picture)
	saved, inserted, err := uc.userRepo.Save(ctx, newUser)
	if err != nil {
		return nil, err
	}
	if !inserted {
		// 同時リクエストの相手が先にこのユーザーを作成済み。通知設定は
		// 相手側の処理で作られるため、ここでは重複作成しない。
		return saved, nil
	}

	// 新規ユーザーにはデフォルトの通知設定（メール通知ON）を作成する
	setting := domainNotification.NewSetting(uc.idGenerator.NewNotificationSettingID(), saved.ID())
	if err := uc.notificationRepo.Create(ctx, setting); err != nil {
		return nil, err
	}

	return saved, nil
}
