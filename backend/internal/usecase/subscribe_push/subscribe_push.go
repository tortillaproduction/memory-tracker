package subscribe_push

import (
	"context"

	"github.com/tortillaproduction/memory-tracker/internal/domain/notification"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

type IDGenerator interface {
	NewPushSubscriptionID() notification.SubscriptionID
}

type Input struct {
	UserID    user.ID
	Endpoint  string
	P256dhKey string
	AuthKey   string
	UserAgent string
}

type Usecase struct {
	pushSubRepo notification.PushSubscriptionRepository
	settingRepo notification.SettingRepository
	idGenerator IDGenerator
}

func NewUsecase(pushSubRepo notification.PushSubscriptionRepository, settingRepo notification.SettingRepository, idGen IDGenerator) *Usecase {
	return &Usecase{pushSubRepo: pushSubRepo, settingRepo: settingRepo, idGenerator: idGen}
}

// Execute は購読情報を保存し(同一endpointなら上書き)、そのユーザーのpush_enabledを
// trueにする。これによりバッチはこのユーザーをプッシュ優先チャネルとして扱うようになる。
func (uc *Usecase) Execute(ctx context.Context, in Input) error {
	sub := notification.NewPushSubscription(
		uc.idGenerator.NewPushSubscriptionID(),
		in.UserID,
		in.Endpoint,
		in.P256dhKey,
		in.AuthKey,
		in.UserAgent,
	)
	if err := uc.pushSubRepo.Create(ctx, sub); err != nil {
		return err
	}

	setting, err := uc.settingRepo.FindByUserID(ctx, in.UserID)
	if err != nil {
		return err
	}
	if setting.PushEnabled() {
		return nil
	}
	setting.SetPushEnabled(true)
	return uc.settingRepo.Update(ctx, setting)
}
