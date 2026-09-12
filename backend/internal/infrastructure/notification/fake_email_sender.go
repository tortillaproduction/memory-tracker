package notification

// SentEmail は FakeEmailSender が記録した1通分の送信内容。
type SentEmail struct {
	To      string
	ToName  string
	Subject string
	Body    string
}

// FakeEmailSender は実際の送信を行わず、呼び出し内容をメモリ上に記録するだけの
// EmailSender実装。テストで通知バッチの挙動を検証するために使う。
type FakeEmailSender struct {
	Sent []SentEmail
	// SendErr が設定されている場合、Send はこのエラーを返し、記録も行わない。
	SendErr error
}

func NewFakeEmailSender() *FakeEmailSender {
	return &FakeEmailSender{}
}

func (s *FakeEmailSender) Send(to, toName, subject, body string) error {
	if s.SendErr != nil {
		return s.SendErr
	}

	s.Sent = append(s.Sent, SentEmail{To: to, ToName: toName, Subject: subject, Body: body})
	return nil
}
