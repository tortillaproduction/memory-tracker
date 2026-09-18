package unsubscribe_push

import (
	"context"

	"github.com/tortillaproduction/memory-tracker/internal/domain/notification"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

type Input struct {
	UserID   user.ID
	Endpoint string
}

type Usecase struct {
	pushSubRepo notification.PushSubscriptionRepository
	settingRepo notification.SettingRepository
}

func NewUsecase(pushSubRepo notification.PushSubscriptionRepository, settingRepo notification.SettingRepository) *Usecase {
	return &Usecase{pushSubRepo: pushSubRepo, settingRepo: settingRepo}
}

// Execute はendpointがそのユーザー自身の購読であることを確認した上で削除し、
// 残りの購読が0件になったらpush_enabledをfalseに戻す。
func (uc *Usecase) Execute(ctx context.Context, in Input) error {
	subs, err := uc.pushSubRepo.ListByUserID(ctx, in.UserID)
	if err != nil {
		return err
	}

	owned := false
	for _, s := range subs {
		if s.Endpoint() == in.Endpoint {
			owned = true
			break
		}
	}
	if !owned {
		return nil
	}

	if err := uc.pushSubRepo.DeleteByEndpoint(ctx, in.Endpoint); err != nil {
		return err
	}

	if len(subs) > 1 {
		// まだ他の購読が残っている
		return nil
	}

	setting, err := uc.settingRepo.FindByUserID(ctx, in.UserID)
	if err != nil {
		return err
	}
	setting.SetPushEnabled(false)
	return uc.settingRepo.Update(ctx, setting)
}
