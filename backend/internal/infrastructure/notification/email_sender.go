package notification

import (
	"fmt"

	"github.com/resend/resend-go/v2"
)

// EmailSender はResend経由でメールを送信する。
// インターフェースとして切り出しているので、テスト時はモックに差し替えられる。
// htmlBody が空文字列の場合はテキストメールとして送信する。
type EmailSender interface {
	Send(to, toName, subject, textBody, htmlBody string) error
}

// noopEmailSender はRESEND_API_KEY未設定時に使う何もしない実装。
// プッシュ通知だけで運用する環境や開発時でも、通知バッチ自体は起動できるようにする。
type noopEmailSender struct{}

func NewNoopEmailSender() EmailSender { return &noopEmailSender{} }

func (s *noopEmailSender) Send(to, toName, subject, textBody, htmlBody string) error { return nil }

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

func (s *resendEmailSender) Send(to, toName, subject, textBody, htmlBody string) error {
	from := fmt.Sprintf("%s <%s>", s.fromName, s.fromAddress)

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Text:    textBody,
		Html:    htmlBody,
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("resend send: %w", err)
	}

	return nil
}
