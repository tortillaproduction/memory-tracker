package notification

import (
	"context"
	"errors"

	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

var ErrSettingNotFound = errors.New("notification setting not found")

type SettingRepository interface {
	Create(ctx context.Context, s *Setting) error
	FindByUserID(ctx context.Context, userID user.ID) (*Setting, error)
	Update(ctx context.Context, s *Setting) error
}

// PushSubscriptionRepository はプッシュ購読の永続化を担う。
// DeleteByEndpointはバッチのセルフヒーリング(410 Gone時の自動削除)、
// DeleteByIDは設定画面からの明示的な解除で使う。
type PushSubscriptionRepository interface {
	Create(ctx context.Context, sub *PushSubscription) error
	DeleteByEndpoint(ctx context.Context, endpoint string) error
	DeleteByID(ctx context.Context, userID user.ID, id SubscriptionID) error
	ListByUserID(ctx context.Context, userID user.ID) ([]*PushSubscription, error)
}
