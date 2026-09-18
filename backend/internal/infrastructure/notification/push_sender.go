package notification

import (
	"errors"
	"fmt"
	"net/http"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// ErrSubscriptionGone は購読が失効している(404/410)ことを示す。呼び出し側は
// このエラーを受け取ったら該当の購読情報を削除し、次サイクル以降のプッシュ送信を
// 諦めてメール通知にフォールバックする(セルフヒーリング)。
var ErrSubscriptionGone = errors.New("push subscription is gone")

// PushSender はブラウザのPush APIにWeb Pushメッセージを送信する。
// EmailSenderと対になるインターフェースで、テスト時はFakePushSenderに差し替える。
type PushSender interface {
	Send(sub PushSubscriptionTarget, payloadJSON []byte) error
}

// PushSubscriptionTarget は送信に必要な購読情報の最小集合。
// domain/notification.PushSubscriptionへの直接依存を避けるためこのパッケージ内で定義する。
type PushSubscriptionTarget struct {
	Endpoint  string
	P256dhKey string
	AuthKey   string
}

// noopPushSender はVAPID鍵未設定時に使う何もしない実装。
// メール専用で運用する既存環境でも、購読/設定APIやバッチ自体は安全に動作させる。
type noopPushSender struct{}

func NewNoopPushSender() PushSender { return &noopPushSender{} }

func (s *noopPushSender) Send(sub PushSubscriptionTarget, payloadJSON []byte) error {
	// VAPID鍵未設定であることを示す一時的エラー。ErrSubscriptionGoneではないため
	// 購読は削除されず、Executeはメール通知にフォールバックする。
	return errors.New("push notifications are not configured (VAPID keys missing)")
}

type webPushSender struct {
	vapidPublicKey  string
	vapidPrivateKey string
	subscriber      string
}

func NewWebPushSender(vapidPublicKey, vapidPrivateKey, subscriber string) PushSender {
	return &webPushSender{
		vapidPublicKey:  vapidPublicKey,
		vapidPrivateKey: vapidPrivateKey,
		subscriber:      subscriber,
	}
}

func (s *webPushSender) Send(sub PushSubscriptionTarget, payloadJSON []byte) error {
	resp, err := webpush.SendNotification(payloadJSON, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dhKey,
			Auth:   sub.AuthKey,
		},
	}, &webpush.Options{
		Subscriber:      s.subscriber,
		VAPIDPublicKey:  s.vapidPublicKey,
		VAPIDPrivateKey: s.vapidPrivateKey,
		TTL:             60 * 60, // 1時間。ユーザーが長時間オフラインでも古い期限切れ通知を後追いで届けすぎない範囲。
	})
	if err != nil {
		return fmt.Errorf("webpush send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		return ErrSubscriptionGone
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webpush send: unexpected status %d", resp.StatusCode)
	}

	return nil
}
