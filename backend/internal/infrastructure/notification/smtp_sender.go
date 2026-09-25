package notification

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"time"
)

const smtpTimeout = 30 * time.Second

// smtpEmailSender はSTARTTLS対応のSMTPサーバー(Gmailなど)経由でメールを送信する。
type smtpEmailSender struct {
	host        string
	port        string
	username    string
	password    string
	fromAddress string
	fromName    string
}

func NewSMTPEmailSender(host, port, username, password, fromAddress, fromName string) EmailSender {
	return &smtpEmailSender{
		host:        host,
		port:        port,
		username:    username,
		password:    password,
		fromAddress: fromAddress,
		fromName:    fromName,
	}
}

func (s *smtpEmailSender) Send(to, toName, subject, textBody, htmlBody string) error {
	msg, err := buildMIMEMessage(
		mail.Address{Name: s.fromName, Address: s.fromAddress},
		mail.Address{Name: toName, Address: to},
		subject, textBody, htmlBody,
	)
	if err != nil {
		return fmt.Errorf("smtp build message: %w", err)
	}

	// smtp.SendMailはタイムアウトを持たないため、通知バッチが止まらないよう接続に期限を付ける。
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(s.host, s.port), smtpTimeout)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(smtpTimeout)); err != nil {
		return fmt.Errorf("smtp set deadline: %w", err)
	}

	c, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer c.Close()

	if err := c.StartTLS(&tls.Config{ServerName: s.host}); err != nil {
		return fmt.Errorf("smtp starttls: %w", err)
	}
	if err := c.Auth(smtp.PlainAuth("", s.username, s.password, s.host)); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := c.Mail(s.fromAddress); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp data close: %w", err)
	}
	return c.Quit()
}

func buildMIMEMessage(from, to mail.Address, subject, textBody, htmlBody string) ([]byte, error) {
	var buf bytes.Buffer
	writeHeader := func(k, v string) { fmt.Fprintf(&buf, "%s: %s\r\n", k, v) }

	writeHeader("From", from.String())
	writeHeader("To", to.String())
	writeHeader("Subject", mime.BEncoding.Encode("UTF-8", subject))
	writeHeader("Date", time.Now().Format(time.RFC1123Z))
	writeHeader("MIME-Version", "1.0")

	if htmlBody == "" {
		writeHeader("Content-Type", `text/plain; charset="UTF-8"`)
		writeHeader("Content-Transfer-Encoding", "base64")
		buf.WriteString("\r\n")
		writeBase64(&buf, textBody)
		return buf.Bytes(), nil
	}

	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	boundary := "mt-" + hex.EncodeToString(b)

	writeHeader("Content-Type", fmt.Sprintf(`multipart/alternative; boundary="%s"`, boundary))
	buf.WriteString("\r\n")
	for _, part := range []struct{ ctype, body string }{
		{"text/plain", textBody},
		{"text/html", htmlBody},
	} {
		fmt.Fprintf(&buf, "--%s\r\n", boundary)
		fmt.Fprintf(&buf, "Content-Type: %s; charset=\"UTF-8\"\r\n", part.ctype)
		buf.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
		writeBase64(&buf, part.body)
	}
	fmt.Fprintf(&buf, "--%s--\r\n", boundary)
	return buf.Bytes(), nil
}

// RFC 2045 の1行76文字制限に合わせて折り返す。
func writeBase64(buf *bytes.Buffer, s string) {
	enc := base64.StdEncoding.EncodeToString([]byte(s))
	for len(enc) > 76 {
		buf.WriteString(enc[:76])
		buf.WriteString("\r\n")
		enc = enc[76:]
	}
	buf.WriteString(enc)
	buf.WriteString("\r\n")
}
