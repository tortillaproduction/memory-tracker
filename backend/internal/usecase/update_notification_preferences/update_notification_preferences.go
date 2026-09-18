package update_notification_preferences

import (
	"context"

	"github.com/tortillaproduction/memory-tracker/internal/domain/notification"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

// Preferences は設定画面に表示・更新する通知設定のDTO。
type Preferences struct {
	EmailEnabled                  bool
	PushEnabled                   bool
	DisableEmailWhenPushAvailable bool
}

type Usecase struct {
	settingRepo notification.SettingRepository
}

func NewUsecase(settingRepo notification.SettingRepository) *Usecase {
	return &Usecase{settingRepo: settingRepo}
}

func (uc *Usecase) Get(ctx context.Context, userID user.ID) (Preferences, error) {
	setting, err := uc.settingRepo.FindByUserID(ctx, userID)
	if err != nil {
		return Preferences{}, err
	}
	return toPreferences(setting), nil
}

// UpdateDisableEmailWhenPushAvailable は「プッシュが使える場合はメールを止める」設定のみを更新する。
func (uc *Usecase) UpdateDisableEmailWhenPushAvailable(ctx context.Context, userID user.ID, value bool) (Preferences, error) {
	setting, err := uc.settingRepo.FindByUserID(ctx, userID)
	if err != nil {
		return Preferences{}, err
	}
	setting.SetDisableEmailWhenPushAvailable(value)
	if err := uc.settingRepo.Update(ctx, setting); err != nil {
		return Preferences{}, err
	}
	return toPreferences(setting), nil
}

func toPreferences(s *notification.Setting) Preferences {
	return Preferences{
		EmailEnabled:                  s.EmailEnabled(),
		PushEnabled:                   s.PushEnabled(),
		DisableEmailWhenPushAvailable: s.DisableEmailWhenPushAvailable(),
	}
}
