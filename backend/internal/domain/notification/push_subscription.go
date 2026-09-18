package notification

import "github.com/tortillaproduction/memory-tracker/internal/domain/user"

type SubscriptionID string

// PushSubscription はブラウザのPush APIが発行した購読情報。1ユーザーが複数デバイスで
// 購読すると複数行になる(端末ごとにendpointが異なるため)。
type PushSubscription struct {
	id        SubscriptionID
	userID    user.ID
	endpoint  string
	p256dhKey string
	authKey   string
	userAgent string
}

func NewPushSubscription(id SubscriptionID, userID user.ID, endpoint, p256dhKey, authKey, userAgent string) *PushSubscription {
	return &PushSubscription{
		id:        id,
		userID:    userID,
		endpoint:  endpoint,
		p256dhKey: p256dhKey,
		authKey:   authKey,
		userAgent: userAgent,
	}
}

func (p *PushSubscription) ID() SubscriptionID { return p.id }
func (p *PushSubscription) UserID() user.ID    { return p.userID }
func (p *PushSubscription) Endpoint() string   { return p.endpoint }
func (p *PushSubscription) P256dhKey() string  { return p.p256dhKey }
func (p *PushSubscription) AuthKey() string    { return p.authKey }
func (p *PushSubscription) UserAgent() string  { return p.userAgent }
