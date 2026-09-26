// Package push_delivery は「Web Pushを送ったが、端末側で実際に表示されたか」を追跡する。
//
// プッシュサービスが返す2xxは「受理した」だけで、ブラウザ/OSが通知を表示したかは
// サーバーからは分からない。そこでService Workerに表示結果を報告させ(POST /api/push/ack)、
// 一定時間報告が無い送信を警告ログに残す。
package push_delivery

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"sync"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

// AckTimeout は送信後この時間を過ぎても報告が来ない場合に「未確認」として警告する。
const AckTimeout = 10 * time.Minute

// tokenTTL はプッシュのTTL(1時間)に合わせる。端末がオフラインで後から受信した場合も報告できる。
const tokenTTL = time.Hour + 10*time.Minute

const (
	StatusShown  = "shown"  // showNotificationが成功した
	StatusFailed = "failed" // 権限なし・showNotification失敗など
)

// TokenIssuer は報告用トークンを発行・検証する(auth.PushAckTokenIssuerが満たす)。
type TokenIssuer interface {
	IssuePushAckToken(notificationID string, userID user.ID, expiresAt time.Time) string
	VerifyPushAckToken(token string) (string, user.ID, error)
}

// Report はService Workerからの表示報告。
type Report struct {
	Token      string `json:"token"`
	Status     string `json:"status"`
	Permission string `json:"permission"` // SW内のNotification.permission
	Error      string `json:"error"`
}

type pending struct {
	userID   user.ID
	host     string
	sentAt   time.Time
	notified bool
}

type Tracker struct {
	issuer  TokenIssuer
	logger  *slog.Logger
	mu      sync.Mutex
	pending map[string]*pending
}

func NewTracker(issuer TokenIssuer, logger *slog.Logger) *Tracker {
	return &Tracker{issuer: issuer, logger: logger, pending: make(map[string]*pending)}
}

// Track は1回のプッシュ送信を追跡対象に登録し、ペイロードに埋め込む報告トークンを返す。
// pushHostはプッシュサービスのホスト名(FCM/Mozilla等の切り分け用。endpoint全体は秘匿情報のため使わない)。
func (t *Tracker) Track(userID user.ID, pushHost string, now time.Time) (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	id := hex.EncodeToString(buf)

	t.mu.Lock()
	t.pending[id] = &pending{userID: userID, host: pushHost, sentAt: now}
	t.mu.Unlock()

	return t.issuer.IssuePushAckToken(id, userID, now.Add(tokenTTL)), nil
}

// Forget は送信自体に失敗した場合に追跡対象から外す(未確認警告を出さないため)。
func (t *Tracker) Forget(token string) {
	id, _, err := t.issuer.VerifyPushAckToken(token)
	if err != nil {
		return
	}
	t.mu.Lock()
	delete(t.pending, id)
	t.mu.Unlock()
}

// Ack は表示報告を受け付けてログに残す。トークンが不正な場合はfalseを返す。
func (t *Tracker) Ack(r Report, now time.Time) bool {
	id, userID, err := t.issuer.VerifyPushAckToken(r.Token)
	if err != nil {
		return false
	}

	t.mu.Lock()
	p, tracked := t.pending[id]
	delete(t.pending, id)
	t.mu.Unlock()

	attrs := []any{"userID", userID, "notificationID", id, "permission", r.Permission}
	if tracked {
		attrs = append(attrs, "pushHost", p.host, "latency", now.Sub(p.sentAt).Round(time.Second))
		if p.notified {
			attrs = append(attrs, "late", true) // 未確認警告を出した後に届いた報告
		}
	}

	if r.Status == StatusShown {
		t.logger.Info("push notification displayed on client", attrs...)
		return true
	}

	reason := "unknown"
	switch {
	case r.Permission != "" && r.Permission != "granted":
		reason = "notification permission is " + r.Permission + " (browser/OS setting)"
	case r.Error != "":
		reason = r.Error
	}
	t.logger.Error("push notification was not displayed on client", append(attrs, "reason", reason)...)
	return true
}

// Sweep は報告が来ないまま AckTimeout を過ぎた送信を警告ログに出す。
// ブラウザ未起動/オフライン、OSの集中モード等、サーバーから区別できない原因を含む。
// 遅れて報告が届く可能性があるため、警告後もTTL相当の間は追跡を続ける。
func (t *Tracker) Sweep(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for id, p := range t.pending {
		age := now.Sub(p.sentAt)
		if age > tokenTTL {
			delete(t.pending, id)
			continue
		}
		if !p.notified && age > AckTimeout {
			p.notified = true
			t.logger.Warn("push notification accepted by push service but not acknowledged by client",
				"userID", p.userID, "notificationID", id, "pushHost", p.host, "elapsed", age.Round(time.Second),
				"hint", "browser may be closed/offline, notification permission revoked, or OS-level notification suppression")
		}
	}
}
