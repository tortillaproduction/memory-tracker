// Package dbtest は実DB(Postgres)を使う統合テストの共通処理をまとめる。
// ローカル実行前に `docker compose up -d db` でDBを起動し、マイグレーションを適用しておくこと。
// DATABASE_URL未設定時は docker-compose.yml のローカル既定値に接続する。
package dbtest

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	. "github.com/onsi/gomega"
)

const defaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/memorytracker?sslmode=disable"

// Open はテスト用DBに接続する。接続できない場合は(nil, false)を返し、呼び出し側でスキップする。
func Open() (*sql.DB, bool) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = defaultDatabaseURL
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, false
	}
	return db, true
}

func NewID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, ulid.Make().String())
}

// CreateUser はユーザーと通知設定(メール有効、emailモード)を作成し、userIDを返す。
func CreateUser(ctx context.Context, db *sql.DB) string {
	userID := NewID("user")
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, google_id, email, name)
		VALUES ($1, $2, $3, $4)
	`, userID, NewID("google"), userID+"@example.com", "Test User")
	Expect(err).NotTo(HaveOccurred())

	_, err = db.ExecContext(ctx, `
		INSERT INTO notification_settings (id, user_id, email_enabled)
		VALUES ($1, $2, true)
	`, NewID("ns"), userID)
	Expect(err).NotTo(HaveOccurred())
	return userID
}

func CreateSite(ctx context.Context, db *sql.DB, userID, name string, intervalHours int) string {
	siteID := NewID("site")
	_, err := db.ExecContext(ctx, `
		INSERT INTO sites (id, user_id, name, url, interval_hours)
		VALUES ($1, $2, $3, $4, $5)
	`, siteID, userID, name, "https://example.com/"+siteID, intervalHours)
	Expect(err).NotTo(HaveOccurred())
	return siteID
}

func ArchiveSite(ctx context.Context, db *sql.DB, siteID string) {
	_, err := db.ExecContext(ctx, `UPDATE sites SET is_archived = true WHERE id = $1`, siteID)
	Expect(err).NotTo(HaveOccurred())
}

func InsertCheckIn(ctx context.Context, db *sql.DB, userID, siteID string, checkedAt time.Time, isInitial bool) {
	_, err := db.ExecContext(ctx, `
		INSERT INTO check_ins (id, user_id, site_id, checked_at, is_initial)
		VALUES ($1, $2, $3, $4, $5)
	`, NewID("checkin"), userID, siteID, checkedAt, isInitial)
	Expect(err).NotTo(HaveOccurred())
}

func InsertNotificationLog(ctx context.Context, db *sql.DB, userID, siteID string, sentAt time.Time) {
	_, err := db.ExecContext(ctx, `
		INSERT INTO notification_logs (id, user_id, site_id, sent_at)
		VALUES ($1, $2, $3, $4)
	`, NewID("nlog"), userID, siteID, sentAt)
	Expect(err).NotTo(HaveOccurred())
}

func CountCheckIns(ctx context.Context, db *sql.DB, siteID string) int {
	var n int
	Expect(db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM check_ins WHERE site_id = $1 AND is_initial = false`, siteID,
	).Scan(&n)).To(Succeed())
	return n
}

// CleanupUser はユーザーを削除する。ON DELETE CASCADEで関連行もすべて消える。
func CleanupUser(ctx context.Context, db *sql.DB, userID string) {
	_, _ = db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, userID)
}
