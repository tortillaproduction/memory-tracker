package notification

// SentPush は FakePushSender が記録した1件分の送信内容。
type SentPush struct {
	Endpoint string
	Payload  []byte
}

// FakePushSender は実際の送信を行わず、呼び出し内容をメモリ上に記録するだけの
// PushSender実装。テストで通知バッチの挙動を検証するために使う。
type FakePushSender struct {
	Sent []SentPush
	// GoneEndpoints に含まれるendpointへのSendはErrSubscriptionGoneを返す。
	GoneEndpoints map[string]bool
	// SendErr が設定されている場合、GoneEndpoints判定より先にこのエラーを返す。
	SendErr error
}

func NewFakePushSender() *FakePushSender {
	return &FakePushSender{GoneEndpoints: map[string]bool{}}
}

func (s *FakePushSender) Send(sub PushSubscriptionTarget, payloadJSON []byte) error {
	if s.SendErr != nil {
		return s.SendErr
	}
	if s.GoneEndpoints[sub.Endpoint] {
		return ErrSubscriptionGone
	}

	s.Sent = append(s.Sent, SentPush{Endpoint: sub.Endpoint, Payload: payloadJSON})
	return nil
}
