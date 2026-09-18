package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/site"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

var ErrInvalidCheckinToken = errors.New("invalid checkin token")

// CheckinTokenIssuer はメール経由の/go/{siteId}リンクに埋め込む、Cookie不要の
// ワンタイムトークンを発行・検証する。スマホのメールアプリはアプリ内WebViewで
// リンクを開くことが多く、ログイン中のブラウザとCookieが共有されないため、
// セッションCookieに依存しない認証手段としてこれを使う。
// サーバー側に状態を持たない(自己検証)ので、失効させたい場合は有効期限で対応する。
type CheckinTokenIssuer struct {
	secret []byte
}

func NewCheckinTokenIssuer(secret string) *CheckinTokenIssuer {
	return &CheckinTokenIssuer{secret: []byte(secret)}
}

// IssueCheckinToken はsiteID+userID+有効期限を署名したトークンを発行する。
func (i *CheckinTokenIssuer) IssueCheckinToken(userID user.ID, siteID site.ID, expiresAt time.Time) (string, error) {
	payload := encodePayload(siteID, userID, expiresAt)
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))
	sig := i.sign(encodedPayload)
	return encodedPayload + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

// VerifyCheckinToken はトークンの署名・有効期限・siteIDの一致を検証し、
// 問題なければ埋め込まれていたuserIDを返す。
func (i *CheckinTokenIssuer) VerifyCheckinToken(siteID site.ID, token string) (user.ID, error) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return "", ErrInvalidCheckinToken
	}
	encodedPayload, encodedSig := parts[0], parts[1]

	sig, err := base64.RawURLEncoding.DecodeString(encodedSig)
	if err != nil {
		return "", ErrInvalidCheckinToken
	}
	if !hmac.Equal(sig, i.sign(encodedPayload)) {
		return "", ErrInvalidCheckinToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return "", ErrInvalidCheckinToken
	}

	tokenSiteID, userID, expiresAt, err := decodePayload(string(payloadBytes))
	if err != nil {
		return "", ErrInvalidCheckinToken
	}
	if tokenSiteID != siteID {
		return "", ErrInvalidCheckinToken
	}
	if time.Now().After(expiresAt) {
		return "", ErrInvalidCheckinToken
	}

	return userID, nil
}

func (i *CheckinTokenIssuer) sign(encodedPayload string) []byte {
	mac := hmac.New(sha256.New, i.secret)
	mac.Write([]byte(encodedPayload))
	return mac.Sum(nil)
}

func encodePayload(siteID site.ID, userID user.ID, expiresAt time.Time) string {
	return fmt.Sprintf("%s|%s|%d", siteID, userID, expiresAt.Unix())
}

func decodePayload(payload string) (site.ID, user.ID, time.Time, error) {
	parts := strings.SplitN(payload, "|", 3)
	if len(parts) != 3 {
		return "", "", time.Time{}, ErrInvalidCheckinToken
	}
	expiresAtUnix, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return "", "", time.Time{}, ErrInvalidCheckinToken
	}
	return site.ID(parts[0]), user.ID(parts[1]), time.Unix(expiresAtUnix, 0), nil
}
