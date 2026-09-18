package notify_overdue_sites

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	domainNotification "github.com/tortillaproduction/memory-tracker/internal/domain/notification"
	"github.com/tortillaproduction/memory-tracker/internal/domain/site"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
	"github.com/tortillaproduction/memory-tracker/internal/infrastructure/notification"
)

// checkinLinkTTL はメールに埋め込むチェックインリンクの有効期限。
// この期間を過ぎたリンクを踏んでもチェックインは記録されない。
const checkinLinkTTL = 7 * 24 * time.Hour

// SiteRow はバッチで使うサイト情報の最小DTO。
type SiteRow struct {
	SiteID                        string
	SiteName                      string
	SiteURL                       string
	IntervalHours                 int
	UserID                        string
	UserEmail                     string
	UserName                      string
	LastCheckedAt                 *time.Time // is_initial=falseの最終チェックイン。nilは一度も開いていない。
	CheckinURL                    string     // /go/{siteId}?token=... 送信直前にセットする
	EmailEnabled                  bool
	DisableEmailWhenPushAvailable bool
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
	pushSender  notification.PushSender
	pushSubRepo domainNotification.PushSubscriptionRepository
	settingRepo domainNotification.SettingRepository
	logger      *slog.Logger
	frontendURL string
	tokenIssuer TokenIssuer
}

func NewUsecase(
	db *sql.DB,
	emailSender notification.EmailSender,
	pushSender notification.PushSender,
	pushSubRepo domainNotification.PushSubscriptionRepository,
	settingRepo domainNotification.SettingRepository,
	logger *slog.Logger,
	frontendURL string,
	tokenIssuer TokenIssuer,
) *Usecase {
	return &Usecase{
		db:          db,
		emailSender: emailSender,
		pushSender:  pushSender,
		pushSubRepo: pushSubRepo,
		settingRepo: settingRepo,
		logger:      logger,
		frontendURL: frontendURL,
		tokenIssuer: tokenIssuer,
	}
}

// Execute は全ユーザーの全サイトを確認し、期限切れかつ未通知のサイトがあれば通知する。
// 通知チャネルの決定ポリシー(自動切替+セルフヒーリング)はnotifyUserを参照。
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
		if err := uc.notifyUser(ctx, sites, now); err != nil {
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
			latest_ci.checked_at,
			ns.email_enabled,
			ns.disable_email_when_push_available
		FROM sites s
		JOIN users u ON u.id = s.user_id
		JOIN notification_settings ns ON ns.user_id = u.id
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
			-- メールかプッシュのどちらか一方でも有効なユーザーのみ。
			-- 実際にどちらのチャネルで送るかはnotifyUserがプッシュ購読の生死を見て都度判定する。
			AND (ns.email_enabled = true OR ns.push_enabled = true)
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
			&r.EmailEnabled, &r.DisableEmailWhenPushAvailable,
		); err != nil {
			return nil, err
		}
		result = append(result, r)
	}

	return result, dbRows.Err()
}

// notifyUser は1ユーザー分の期限切れサイトをまとめて通知する。
//
// チャネル決定ポリシー(自動切替+セルフヒーリング):
//   - 有効なプッシュ購読があればまずプッシュを試す。410 Goneが返った購読は
//     即座に削除し(自己修復)、全滅した場合は同じサイクル内でメールにフォールバックする。
//   - メールは「emailEnabledかつ(プッシュが1件も届かなかった、またはdisableEmailWhenPushAvailableがfalse)」
//     の場合にのみ送る。
func (uc *Usecase) notifyUser(ctx context.Context, sites []*SiteRow, now time.Time) error {
	owner := sites[0]

	if err := uc.issueCheckinLinks(sites, now); err != nil {
		return fmt.Errorf("issue checkin links: %w", err)
	}

	pushOK := uc.sendPush(ctx, owner, sites, now)

	shouldEmail := owner.EmailEnabled && (!pushOK || !owner.DisableEmailWhenPushAvailable)

	if shouldEmail {
		if err := uc.sendEmail(owner, sites, now); err != nil {
			return err
		}
	}

	if !pushOK && !shouldEmail {
		// どちらのチャネルにも送れなかった(例: プッシュ購読が全滅し、メールも無効)。
		// notification_logsには記録せず、次サイクルで再度対象にする。
		return nil
	}

	return uc.saveNotificationLogs(ctx, sites, now)
}

// sendPush はユーザーの全プッシュ購読にWeb Pushを送る。1件でも届けば true を返す。
// 410 Goneが返った購読はその場で削除し(セルフヒーリング)、全購読が消滅した場合は
// notification_settings.push_enabledもfalseに戻す。
func (uc *Usecase) sendPush(ctx context.Context, owner *SiteRow, sites []*SiteRow, now time.Time) bool {
	subs, err := uc.pushSubRepo.ListByUserID(ctx, user.ID(owner.UserID))
	if err != nil {
		uc.logger.Error("failed to list push subscriptions", "userID", owner.UserID, "error", err)
		return false
	}
	if len(subs) == 0 {
		return false
	}

	payload, err := buildPushPayload(sites, now)
	if err != nil {
		uc.logger.Error("failed to build push payload", "userID", owner.UserID, "error", err)
		return false
	}

	delivered := 0
	remaining := len(subs)
	for _, sub := range subs {
		target := notification.PushSubscriptionTarget{
			Endpoint:  sub.Endpoint(),
			P256dhKey: sub.P256dhKey(),
			AuthKey:   sub.AuthKey(),
		}
		sendErr := uc.pushSender.Send(target, payload)
		switch {
		case sendErr == nil:
			delivered++
		case errors.Is(sendErr, notification.ErrSubscriptionGone):
			if delErr := uc.pushSubRepo.DeleteByEndpoint(ctx, sub.Endpoint()); delErr != nil {
				uc.logger.Error("failed to delete gone push subscription", "userID", owner.UserID, "error", delErr)
			}
			remaining--
		default:
			// 一時的な失敗とみなし、購読は残して次サイクルの再送に委ねる。
			uc.logger.Error("failed to send push notification", "userID", owner.UserID, "error", sendErr)
		}
	}

	if remaining == 0 {
		if err := uc.disablePushSetting(ctx, user.ID(owner.UserID)); err != nil {
			uc.logger.Error("failed to disable push setting after all subscriptions gone", "userID", owner.UserID, "error", err)
		}
	}

	return delivered > 0
}

// disablePushSetting はプッシュ購読が全滅したユーザーのpush_enabledをfalseに戻す。
// 設定UIの表示を実態に合わせるためで、通知対象クエリ自体はemail_enabledで既にカバーされる。
func (uc *Usecase) disablePushSetting(ctx context.Context, userID user.ID) error {
	setting, err := uc.settingRepo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if !setting.PushEnabled() {
		return nil
	}
	setting.SetPushEnabled(false)
	return uc.settingRepo.Update(ctx, setting)
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
// "not checked yet - every Xh")を組み立てる。メール本文(email_template.go)と
// プッシュ通知のペイロード(push_payload.go)の両方から共有して使う。
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
