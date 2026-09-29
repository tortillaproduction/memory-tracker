package update_notification_mode

import (
	"context"

	"github.com/tortillaproduction/memory-tracker/internal/domain/notification"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

type Usecase struct {
	settingRepo notification.SettingRepository
}

func NewUsecase(settingRepo notification.SettingRepository) *Usecase {
	return &Usecase{settingRepo: settingRepo}
}

func (uc *Usecase) Get(ctx context.Context, userID user.ID) (notification.Mode, error) {
	s, err := uc.settingRepo.FindByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	return s.Mode(), nil
}

// Update は通知モード(email / gate)を切り替える。
func (uc *Usecase) Update(ctx context.Context, userID user.ID, mode notification.Mode) (notification.Mode, error) {
	s, err := uc.settingRepo.FindByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	s.SetMode(mode)
	if err := uc.settingRepo.Update(ctx, s); err != nil {
		return "", err
	}
	return s.Mode(), nil
}
