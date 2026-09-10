package notification

import (
	"fmt"

	"github.com/resend/resend-go/v2"
)

// EmailSender はSendGrid経由でメールを送信する。
// インターフェースとして切り出しているので、テスト時はモックに差し替えられる。
type EmailSender interface {
	Send(to, toName, subject, body string) error
}

type resendEmailSender struct {
	client      *resend.Client
	fromAddress string
	fromName    string
}

func NewResendEmailSender(apiKey, fromAddress, fromName string) EmailSender {
	return &resendEmailSender{
		client:      resend.NewClient(apiKey),
		fromAddress: fromAddress,
		fromName:    fromName,
	}
}

func (s *resendEmailSender) Send(to, toName, subject, body string) error {
	from := fmt.Sprintf("%s <%s>", s.fromName, s.fromAddress)

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Text:    body,
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("resend send: %w", err)
	}

	return nil
}
