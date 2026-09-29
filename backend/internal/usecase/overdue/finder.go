// Package overdue は「期限切れサイト」の読み取りモデルを定義する。
// 期限切れの判定条件はFinderの実装(infrastructure/persistence/postgres)の1か所だけに置き、
// 通知バッチ(notify_overdue_sites)とゲートの候補取得(get_gate_candidates)で共有する。
package overdue

import (
	"context"
	"time"
)

// Site は期限切れサイト1件分。
type Site struct {
	SiteID        string
	SiteName      string
	SiteURL       string
	IntervalHours int
	UserID        string
	UserEmail     string
	UserName      string
	LastCheckedAt *time.Time // is_initial=falseの最終チェックイン。nilは一度も開いていない。
	DueAt         time.Time  // 期限(最終チェックイン、未訪問なら登録時刻 + interval_hours)
}

// OverdueBy は now 時点で期限をどれだけ過ぎているかを返す。
func (s Site) OverdueBy(now time.Time) time.Duration {
	return now.Sub(s.DueAt)
}

// Filter は期限切れ条件に加える絞り込み。ゼロ値は「全ユーザーの全期限切れサイト」。
type Filter struct {
	// UserID が空でなければ、そのユーザーのサイトだけに絞る。
	UserID string
	// ExcludeRecentlyNotified は、前回の通知からinterval_hoursが経っていないサイトを除く
	// (通知バッチの二重送信防止)。
	ExcludeRecentlyNotified bool
	// EmailRecipientsOnly は、メール通知の送信対象ユーザーだけに絞る。
	// email_enabledで、かつ「メールモード」または「ゲートモードだがゲート端末が
	// GateActiveSince以降にAPIを使っていない(未接続とみなす)」ユーザーが対象。
	EmailRecipientsOnly bool
	GateActiveSince     time.Time
	// Limit が正なら、期限を大きく過ぎている順に最大Limit件を返す。
	Limit int
}

// Finder は期限切れサイトを期限の古い順(超過時間の降順)で返す。
type Finder interface {
	ListOverdue(ctx context.Context, now time.Time, f Filter) ([]Site, error)
}
