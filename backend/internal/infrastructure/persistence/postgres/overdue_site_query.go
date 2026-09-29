package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/usecase/overdue"
)

type overdueSiteQuery struct {
	db *sql.DB
}

func NewOverdueSiteQuery(db *sql.DB) overdue.Finder {
	return &overdueSiteQuery{db: db}
}

// ListOverdue は期限切れのサイトを返す。
//
// 期限切れの定義(この関数だけに置く):
//   - チェックイン済み: 最終チェックイン(is_initial=false)からinterval_hours以上経過
//   - 未チェックイン: 登録時の初回チェックイン(is_initial=true)からinterval_hours以上経過
func (q *overdueSiteQuery) ListOverdue(ctx context.Context, now time.Time, f overdue.Filter) ([]overdue.Site, error) {
	args := []any{now}
	arg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	var joins, conds []string
	if f.UserID != "" {
		conds = append(conds, "s.user_id = "+arg(f.UserID))
	}
	if f.EmailRecipientsOnly {
		joins = append(joins, "JOIN notification_settings ns ON ns.user_id = u.id")
		conds = append(conds, `ns.email_enabled = true
			AND (
				ns.mode = 'email'
				-- ゲートモードでも、ゲート端末が一定期間APIを使っていなければ未接続とみなしメールを併用する
				OR NOT EXISTS (
					SELECT 1 FROM gate_tokens gt
					WHERE gt.user_id = u.id AND gt.last_used_at >= `+arg(f.GateActiveSince)+`
				)
			)`)
	}
	if f.ExcludeRecentlyNotified {
		// 二重送信防止: 前回通知からinterval_hours以上経過しているか未通知
		joins = append(joins, `LEFT JOIN LATERAL (
			SELECT sent_at
			FROM notification_logs
			WHERE site_id = s.id
			ORDER BY sent_at DESC
			LIMIT 1
		) latest_nl ON true`)
		conds = append(conds, `(latest_nl.sent_at IS NULL
			OR latest_nl.sent_at < $1::timestamptz - (s.interval_hours || ' hours')::interval)`)
	}

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
			COALESCE(latest_ci.checked_at, initial_ci.checked_at) + (s.interval_hours || ' hours')::interval AS due_at
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
		-- 登録時の初回チェックイン (未チェックインのサイトの期限の基準)
		LEFT JOIN LATERAL (
			SELECT checked_at
			FROM check_ins
			WHERE site_id = s.id AND is_initial = true
			ORDER BY checked_at ASC
			LIMIT 1
		) initial_ci ON true
		` + strings.Join(joins, "\n") + `
		WHERE
			s.is_archived = false
			AND COALESCE(latest_ci.checked_at, initial_ci.checked_at) < $1::timestamptz - (s.interval_hours || ' hours')::interval
	`
	for _, c := range conds {
		query += "\n\t\t\tAND " + c
	}
	query += "\n\t\tORDER BY due_at ASC, s.id"
	if f.Limit > 0 {
		query += "\n\t\tLIMIT " + arg(f.Limit)
	}

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []overdue.Site
	for rows.Next() {
		var s overdue.Site
		if err := rows.Scan(
			&s.SiteID, &s.SiteName, &s.SiteURL, &s.IntervalHours,
			&s.UserID, &s.UserEmail, &s.UserName, &s.LastCheckedAt, &s.DueAt,
		); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}
