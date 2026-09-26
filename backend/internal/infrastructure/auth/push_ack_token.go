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

	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

var ErrInvalidPushAckToken = errors.New("invalid push ack token")

// PushAckTokenIssuer はWeb Pushのペイロードに埋め込む「表示報告」用トークンを発行・検証する。
// Service Workerはセッション Cookie を持たずに(credentials: 'omit')報告を送るため、
// 通知ID+userID+有効期限を署名したトークンを認証代わりに使う。
// チェックイントークンと鍵を共有するが、署名対象に用途のプレフィックスを付けて
// 相互に流用できないようにしている。
type PushAckTokenIssuer struct {
	secret []byte
}

func NewPushAckTokenIssuer(secret string) *PushAckTokenIssuer {
	return &PushAckTokenIssuer{secret: []byte(secret)}
}

func (i *PushAckTokenIssuer) IssuePushAckToken(notificationID string, userID user.ID, expiresAt time.Time) string {
	payload := fmt.Sprintf("%s|%s|%d", notificationID, userID, expiresAt.Unix())
	encoded := base64.RawURLEncoding.EncodeToString([]byte(payload))
	return encoded + "." + base64.RawURLEncoding.EncodeToString(i.sign(encoded))
}

// VerifyPushAckToken は署名と有効期限を検証し、通知IDとuserIDを返す。
func (i *PushAckTokenIssuer) VerifyPushAckToken(token string) (string, user.ID, error) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return "", "", ErrInvalidPushAckToken
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(sig, i.sign(parts[0])) {
		return "", "", ErrInvalidPushAckToken
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", "", ErrInvalidPushAckToken
	}
	fields := strings.SplitN(string(raw), "|", 3)
	if len(fields) != 3 {
		return "", "", ErrInvalidPushAckToken
	}
	exp, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil || time.Now().After(time.Unix(exp, 0)) {
		return "", "", ErrInvalidPushAckToken
	}
	return fields[0], user.ID(fields[1]), nil
}

func (i *PushAckTokenIssuer) sign(encodedPayload string) []byte {
	mac := hmac.New(sha256.New, i.secret)
	mac.Write([]byte("push-ack:"))
	mac.Write([]byte(encodedPayload))
	return mac.Sum(nil)
}
