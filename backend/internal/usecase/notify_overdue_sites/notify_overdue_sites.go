package notify_overdue_sites

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/site"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
	"github.com/tortillaproduction/memory-tracker/internal/infrastructure/notification"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/overdue"
)

// checkinLinkTTL はメールに埋め込むチェックインリンクの有効期限。
// この期間を過ぎたリンクを踏んでもチェックインは記録されない。
const checkinLinkTTL = 7 * 24 * time.Hour

// 通知失敗時の再試行バックオフ。失敗してもnotification_logsには記録されないため、
// 何もしないと毎サイクル(送信先の恒久エラーでも)延々とリトライしてしまう。
// 失敗ごとに待ち時間を倍にし、maxRetryBackoffで頭打ちにする。成功すると解除される。
const (
	baseRetryBackoff = 5 * time.Minute
	maxRetryBackoff  = 6 * time.Hour
)

// gateStaleAfter はゲートモードのユーザーでも、ゲート端末がこの期間APIを使っていなければ
// 未接続(アンインストール・故障など)とみなしてメールを併用するまでの時間。
const gateStaleAfter = 3 * 24 * time.Hour

type retryState struct {
	failures int
	nextTry  time.Time
}

// SiteRow はバッチで使うサイト情報の最小DTO。
type SiteRow struct {
	SiteID        string
	SiteName      string
	SiteURL       string
	IntervalHours int
	UserID        string
	UserEmail     string
	UserName      string
	LastCheckedAt *time.Time // is_initial=falseの最終チェックイン。nilは一度も開いていない。
	CheckinURL    string     // /go/{siteId}?token=... 送信直前にセットする
}

// TokenIssuer はメールのチェックインリンクに埋め込む、Cookie不要のワンタイム
// トークンを発行する。スマホのメールアプリはアプリ内WebViewで開くことが多く、
// ログイン中のブラウザとセッションCookieが共有されないための対応。
type TokenIssuer interface {
	IssueCheckinToken(userID user.ID, siteID site.ID, expiresAt time.Time) (string, error)
}

type Usecase struct {
	db          *sql.DB
	finder      overdue.Finder
	emailSender notification.EmailSender
	logger      *slog.Logger
	frontendURL string
	tokenIssuer TokenIssuer

	mu      sync.Mutex
	retries map[string]*retryState // userID -> 失敗状態(プロセス内のみ保持)
}

func NewUsecase(
	db *sql.DB,
	finder overdue.Finder,
	emailSender notification.EmailSender,
	logger *slog.Logger,
	frontendURL string,
	tokenIssuer TokenIssuer,
) *Usecase {
	return &Usecase{
		db:          db,
		finder:      finder,
		emailSender: emailSender,
		logger:      logger,
		frontendURL: frontendURL,
		tokenIssuer: tokenIssuer,
		retries:     make(map[string]*retryState),
	}
}

// inBackoff は直近の失敗によりまだ再試行を待つべきユーザーかを返す。
func (uc *Usecase) inBackoff(userID string, now time.Time) bool {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	st, ok := uc.retries[userID]
	return ok && now.Before(st.nextTry)
}

// recordFailure は失敗を記録し、次回再試行までの待ち時間(指数バックオフ)を返す。
func (uc *Usecase) recordFailure(userID string, now time.Time) time.Duration {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	st, ok := uc.retries[userID]
	if !ok {
		st = &retryState{}
		uc.retries[userID] = st
	}
	st.failures++
	backoff := baseRetryBackoff
	for i := 1; i < st.failures && backoff < maxRetryBackoff; i++ {
		backoff *= 2
	}
	if backoff > maxRetryBackoff {
		backoff = maxRetryBackoff
	}
	st.nextTry = now.Add(backoff)
	return backoff
}

func (uc *Usecase) clearFailure(userID string) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	delete(uc.retries, userID)
}

// Execute は全ユーザーの全サイトを確認し、期限切れかつ未通知のサイトがあればメールで通知する。
// 二重送信防止: notification_logsの最終送信からinterval_hours以上経過しているサイトのみ対象。
func (uc *Usecase) Execute(ctx context.Context) error {
	now := time.Now()

	rows, err := uc.fetchOverdueSites(ctx, now)
	if err != nil {
		return fmt.Errorf("fetch overdue sites: %w", err)
	}
	if len(rows) == 0 {
		return nil
	}

	// ユーザーごとにグループ化して1回の通知にまとめる
	byUser := make(map[string][]*SiteRow)
	for _, r := range rows {
		byUser[r.UserID] = append(byUser[r.UserID], r)
	}

	for userID, sites := range byUser {
		if uc.inBackoff(userID, now) {
			continue
		}
		if err := uc.notifyUser(ctx, sites, now); err != nil {
			backoff := uc.recordFailure(userID, now)
			uc.logger.Error("failed to send notification", "userID", userID, "error", err, "retryIn", backoff)
			continue // 1ユーザーの失敗で他のユーザーへの通知を止めない
		}
		uc.clearFailure(userID)
		uc.logger.Info("notification sent", "userID", userID, "siteCount", len(sites), "channel", "email")
	}

	return nil
}

// fetchOverdueSites は期限切れかつ今回の通知対象になるサイトを取得する。
// 期限切れの判定はoverdue.Finderに任せ、ここではメールの送信対象と二重送信防止の条件を加える。
func (uc *Usecase) fetchOverdueSites(ctx context.Context, now time.Time) ([]*SiteRow, error) {
	sites, err := uc.finder.ListOverdue(ctx, now, overdue.Filter{
		ExcludeRecentlyNotified: true,
		EmailRecipientsOnly:     true,
		GateActiveSince:         now.Add(-gateStaleAfter),
	})
	if err != nil {
		return nil, err
	}

	result := make([]*SiteRow, 0, len(sites))
	for _, s := range sites {
		result = append(result, &SiteRow{
			SiteID:        s.SiteID,
			SiteName:      s.SiteName,
			SiteURL:       s.SiteURL,
			IntervalHours: s.IntervalHours,
			UserID:        s.UserID,
			UserEmail:     s.UserEmail,
			UserName:      s.UserName,
			LastCheckedAt: s.LastCheckedAt,
		})
	}
	return result, nil
}

// notifyUser は1ユーザー分の期限切れサイトを1通のメールにまとめて通知し、
// 送信できた場合のみnotification_logsに記録する。
func (uc *Usecase) notifyUser(ctx context.Context, sites []*SiteRow, now time.Time) error {
	if err := uc.issueCheckinLinks(sites, now); err != nil {
		return fmt.Errorf("issue checkin links: %w", err)
	}

	if err := uc.sendEmail(sites[0], sites, now); err != nil {
		return err
	}

	return uc.saveNotificationLogs(ctx, sites, now)
}

func (uc *Usecase) sendEmail(owner *SiteRow, sites []*SiteRow, now time.Time) error {
	subject := fmt.Sprintf("::Memory Tracker:: %d site(s) are overdue for a visit", len(sites))

	textBody := buildEmailText(owner.UserName, sites, now)

	htmlBody, err := buildEmailHTML(owner.UserName, sites, now, uc.frontendURL)
	if err != nil {
		// HTML生成に失敗してもテキストメールの送信は継続する
		uc.logger.Error("failed to render HTML email, falling back to text-only", "error", err)
		htmlBody = ""
	}

	return uc.emailSender.Send(owner.UserEmail, owner.UserName, subject, textBody, htmlBody)
}

// issueCheckinLinks は各サイトについてチェックイン用ワンタイムトークンを発行し、
// SiteRow.CheckinURLにセットする。
func (uc *Usecase) issueCheckinLinks(sites []*SiteRow, now time.Time) error {
	expiresAt := now.Add(checkinLinkTTL)
	for _, s := range sites {
		token, err := uc.tokenIssuer.IssueCheckinToken(user.ID(s.UserID), site.ID(s.SiteID), expiresAt)
		if err != nil {
			return err
		}
		s.CheckinURL = fmt.Sprintf("%s/go/%s?token=%s", uc.frontendURL, s.SiteID, url.QueryEscape(token))
	}
	return nil
}

// buildEmailText はHTMLに対応しないメールクライアント向けのプレーンテキスト版本文を組み立てる。
func buildEmailText(userName string, sites []*SiteRow, now time.Time) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Hi %s, \n\n", userName))
	sb.WriteString("The following sites are overdue for a visit.\n")
	sb.WriteString("Click a link to open the site and record your check-in.\n\n")

	for _, s := range sites {
		sb.WriteString(fmt.Sprintf("■ %s\n", s.SiteName))
		if s.LastCheckedAt != nil {
			hours := now.Sub(*s.LastCheckedAt).Hours()
			sb.WriteString(fmt.Sprintf("   Last visited: %.0f hour(s) ago (interval: %d hours)\n", hours, s.IntervalHours))
		} else {
			sb.WriteString(fmt.Sprintf("   Not visited yet (interval: %d hours)\n", s.IntervalHours))
		}

		// s.CheckinURLはワンタイムトークン付きの/go/{siteId}リンク。
		// s.SiteURLを直接貼ると経由せずに開けてしまいチェックインが記録されないため、
		// 必ずこちらを使うこと。
		sb.WriteString(fmt.Sprintf("   -> %s\n\n", s.CheckinURL))
	}

	sb.WriteString("--\nMemory Tracker\n")

	return sb.String()
}

// statusLabel はサイト一行分のステータス文言("last checked Nh ago - every Xh" /
// "not checked yet - every Xh")を組み立てる。メール本文(email_template.go)で使う。
func statusLabel(s *SiteRow, now time.Time) string {
	if s.LastCheckedAt != nil {
		hours := now.Sub(*s.LastCheckedAt).Hours()
		return fmt.Sprintf("last checked %.0fh ago - every %dh", hours, s.IntervalHours)
	}
	return fmt.Sprintf("not checked yet - every %dh", s.IntervalHours)
}

func (uc *Usecase) saveNotificationLogs(ctx context.Context, sites []*SiteRow, now time.Time) error {
	for _, s := range sites {
		id := fmt.Sprintf("%s-%s-%d", s.UserID, s.SiteID, now.UnixNano())
		_, err := uc.db.ExecContext(ctx, `
			INSERT INTO notification_logs (id, user_id, site_id, sent_at)
			VALUES ($1, $2, $3, $4)
		`, id, s.UserID, s.SiteID, now)
		if err != nil {
			return err
		}
	}

	return nil
}
