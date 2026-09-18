package notify_overdue_sites

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/site"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
	"github.com/tortillaproduction/memory-tracker/internal/infrastructure/notification"
)

// checkinLinkTTL はメールに埋め込むチェックインリンクの有効期限。
// この期間を過ぎたリンクを踏んでもチェックインは記録されない。
const checkinLinkTTL = 7 * 24 * time.Hour

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
	emailSender notification.EmailSender
	logger      *slog.Logger
	frontendURL string
	tokenIssuer TokenIssuer
}

func NewUsecase(db *sql.DB, emailSender notification.EmailSender, logger *slog.Logger, frontendURL string, tokenIssuer TokenIssuer) *Usecase {
	return &Usecase{db: db, emailSender: emailSender, logger: logger, frontendURL: frontendURL, tokenIssuer: tokenIssuer}
}

// Execute は全ユーザーの全サイトを確認し、期限切れかつ未通知のサイトがあればメールを送る。
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

	// ユーザーごとにグループ化して1通にまとめる
	byUser := make(map[string][]*SiteRow)
	for _, r := range rows {
		byUser[r.UserID] = append(byUser[r.UserID], r)
	}

	for userID, sites := range byUser {
		if err := uc.sendNotification(ctx, sites, now); err != nil {
			uc.logger.Error("failed to send notification", "userID", userID, "error", err)
			continue // 1ユーザーの失敗で他のユーザーへの通知を止めない
		}
		uc.logger.Info("notification sent", "userID", userID, "siteCount", len(sites))
	}

	return nil
}

// fetchOverdueSites は期限切れかつ今回の通知対象になるサイトを取得する。
func (uc *Usecase) fetchOverdueSites(ctx context.Context, now time.Time) ([]*SiteRow, error) {
	query := `
		SELECT
			s.id,
			s.name,
			s.url,
			s.interval_hours,
			u.id,
			u.email,
			u.name,
			latest_ci.checked_at
		FROM sites s
		JOIN users u ON u.id = s.user_id
		-- is_initial=falseの最終チェックイン (一度も開いていない場合はNULL)
		LEFT JOIN LATERAL (
			SELECT checked_at
			FROM check_ins
			WHERE site_id = s.id AND is_initial = false
			ORDER BY checked_at DESC
			LIMIT 1
		) latest_ci ON true
		-- 直近のnotification_log (二重送信防止に使う)
		LEFT JOIN LATERAL (
			SELECT sent_at
			FROM notification_logs
			WHERE site_id = s.id
			ORDER BY sent_at DESC
			LIMIT 1
		) latest_nl ON true
		WHERE
			s.is_archived = false
			-- メール通知が有効なユーザーのみ
			AND EXISTS (
				SELECT 1 FROM notification_settings ns
				WHERE ns.user_id = u.id AND ns.email_enabled = true
			)
			-- 期限切れ判定:
			--  チェックイン済み -> 最終チェックインからinterval_hours以上経過
			--  未チェックイン -> 登録からinterval_hours以上経過 (is_initial=trueのcheck_in時刻を基準)
			AND (
				(latest_ci.checked_at IS NOT NULL
				  AND latest_ci.checked_at < $1::timestamptz - (s.interval_hours || ' hours')::interval)
				OR
				(latest_ci.checked_at IS NULL
				  AND EXISTS (
				  	SELECT 1 FROM check_ins
					WHERE site_id = s.id AND is_initial = true
					AND checked_at < $1::timestamptz - (s.interval_hours || ' hours')::interval
				  ))
			)
			-- 二重送信防止: 前回通知からinterval_hours以上経過しているか未通知
			AND (
				latest_nl.sent_at IS NULL
				OR latest_nl.sent_at < $1::timestamptz - (s.interval_hours || ' hours')::interval
			)
	`

	dbRows, err := uc.db.QueryContext(ctx, query, now)
	if err != nil {
		return nil, err
	}
	defer dbRows.Close()

	var result []*SiteRow
	for dbRows.Next() {
		r := &SiteRow{}
		if err := dbRows.Scan(
			&r.SiteID, &r.SiteName, &r.SiteURL, &r.IntervalHours,
			&r.UserID, &r.UserEmail, &r.UserName, &r.LastCheckedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, r)
	}

	return result, dbRows.Err()
}

func (uc *Usecase) sendNotification(ctx context.Context, sites []*SiteRow, now time.Time) error {
	owner := sites[0]
	subject := fmt.Sprintf("::Memory Tracker:: %d site(s) are overdue for a visit", len(sites))

	if err := uc.issueCheckinLinks(sites, now); err != nil {
		return fmt.Errorf("issue checkin links: %w", err)
	}

	textBody := buildEmailText(owner.UserName, sites, now)

	htmlBody, err := buildEmailHTML(owner.UserName, sites, now, uc.frontendURL)
	if err != nil {
		// HTML生成に失敗してもテキストメールの送信は継続する
		uc.logger.Error("failed to render HTML email, falling back to text-only", "error", err)
		htmlBody = ""
	}

	if err := uc.emailSender.Send(owner.UserEmail, owner.UserName, subject, textBody, htmlBody); err != nil {
		return err
	}

	// 送信成功後にnotification_logsに記録
	return uc.saveNotificationLogs(ctx, sites, now)
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
