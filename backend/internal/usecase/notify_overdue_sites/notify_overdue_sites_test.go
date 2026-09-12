package notify_overdue_sites_test

// このテストは実DB(Postgres)を必要とする統合テスト。
// ローカル実行前に `docker compose up -d db` でDBを起動しておくこと。
// DATABASE_URL未設定時は docker-compose.yml のローカル既定値に接続する。
// DBに接続できない場合は全テストをスキップする。

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/oklog/ulid/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tortillaproduction/memory-tracker/internal/infrastructure/notification"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/notify_overdue_sites"
)

const defaultTestDatabaseURL = "postgres://postgres:postgres@localhost:5432/memorytracker?sslmode=disable"

var (
	db          *sql.DB
	dbAvailable bool
)

var _ = BeforeSuite(func() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = defaultTestDatabaseURL
	}

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return
	}

	dbAvailable = true
})

var _ = AfterSuite(func() {
	if db != nil {
		db.Close()
	}
})

func newID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, ulid.Make().String())
}

// testUser はユーザー・通知設定を作成し、後片付け用のuserIDを返す。
func createTestUser(ctx context.Context, emailEnabled bool) string {
	userID := newID("user")
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, google_id, email, name)
		VALUES ($1, $2, $3, $4)
	`, userID, newID("google"), userID+"@example.com", "Test User")
	Expect(err).NotTo(HaveOccurred())

	_, err = db.ExecContext(ctx, `
		INSERT INTO notification_settings (id, user_id, email_enabled)
		VALUES ($1, $2, $3)
	`, newID("ns"), userID, emailEnabled)
	Expect(err).NotTo(HaveOccurred())

	return userID
}

func createTestSite(ctx context.Context, userID string, intervalHours int) string {
	siteID := newID("site")
	_, err := db.ExecContext(ctx, `
		INSERT INTO sites (id, user_id, name, url, interval_hours)
		VALUES ($1, $2, $3, $4, $5)
	`, siteID, userID, "Test Site", "https://example.com", intervalHours)
	Expect(err).NotTo(HaveOccurred())

	return siteID
}

func insertCheckIn(ctx context.Context, userID, siteID string, checkedAt time.Time, isInitial bool) {
	_, err := db.ExecContext(ctx, `
		INSERT INTO check_ins (id, user_id, site_id, checked_at, is_initial)
		VALUES ($1, $2, $3, $4, $5)
	`, newID("checkin"), userID, siteID, checkedAt, isInitial)
	Expect(err).NotTo(HaveOccurred())
}

func insertNotificationLog(ctx context.Context, userID, siteID string, sentAt time.Time) {
	_, err := db.ExecContext(ctx, `
		INSERT INTO notification_logs (id, user_id, site_id, sent_at)
		VALUES ($1, $2, $3, $4)
	`, newID("nlog"), userID, siteID, sentAt)
	Expect(err).NotTo(HaveOccurred())
}

func cleanupUser(ctx context.Context, userID string) {
	// usersへのON DELETE CASCADEで関連行(sites, check_ins, notification_settings, notification_logs)も削除される。
	_, _ = db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, userID)
}

var _ = Describe("短期限バッチによる期限切れ通知", func() {
	var (
		ctx    context.Context
		userID string
	)

	BeforeEach(func() {
		if !dbAvailable {
			Skip("DATABASE_URLに接続できないためスキップ (docker compose up -d db を実行してください)")
		}
		ctx = context.Background()
	})

	AfterEach(func() {
		if userID != "" {
			cleanupUser(ctx, userID)
			userID = ""
		}
	})

	It("チェックイン済みで短期限(1時間)が経過している場合、メールが送信される", func() {
		userID = createTestUser(ctx, true)
		siteID := createTestSite(ctx, userID, 1)

		now := time.Now()
		insertCheckIn(ctx, userID, siteID, now.Add(-3*time.Hour), true)  // 初回チェックイン
		insertCheckIn(ctx, userID, siteID, now.Add(-2*time.Hour), false) // 本チェックイン(2時間前、期限は1時間)

		sender := notification.NewFakeEmailSender()
		uc := notify_overdue_sites.NewUsecase(db, sender, slog.Default())

		Expect(uc.Execute(ctx)).To(Succeed())

		Expect(sender.Sent).To(HaveLen(1))
		Expect(sender.Sent[0].Subject).To(ContainSubstring("1 site"))
		Expect(sender.Sent[0].Body).To(ContainSubstring("Test Site"))

		var logCount int
		Expect(db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM notification_logs WHERE site_id = $1`, siteID,
		).Scan(&logCount)).To(Succeed())
		Expect(logCount).To(Equal(1))
	})

	It("一度もチェックインしていないが短期限(1時間)が経過している場合、メールが送信される", func() {
		userID = createTestUser(ctx, true)
		siteID := createTestSite(ctx, userID, 1)

		now := time.Now()
		insertCheckIn(ctx, userID, siteID, now.Add(-2*time.Hour), true) // 初回チェックインのみ、2時間前

		sender := notification.NewFakeEmailSender()
		uc := notify_overdue_sites.NewUsecase(db, sender, slog.Default())

		Expect(uc.Execute(ctx)).To(Succeed())

		Expect(sender.Sent).To(HaveLen(1))
		Expect(sender.Sent[0].Body).To(ContainSubstring("Not visited yet"))
	})

	It("期限内の場合、メールは送信されない", func() {
		userID = createTestUser(ctx, true)
		siteID := createTestSite(ctx, userID, 24)

		now := time.Now()
		insertCheckIn(ctx, userID, siteID, now.Add(-3*time.Hour), true)
		insertCheckIn(ctx, userID, siteID, now.Add(-1*time.Hour), false) // 1時間前、期限は24時間

		sender := notification.NewFakeEmailSender()
		uc := notify_overdue_sites.NewUsecase(db, sender, slog.Default())

		Expect(uc.Execute(ctx)).To(Succeed())
		Expect(sender.Sent).To(BeEmpty())
	})

	It("直近の通知から期限が未経過の場合、二重送信しない", func() {
		userID = createTestUser(ctx, true)
		siteID := createTestSite(ctx, userID, 1)

		now := time.Now()
		insertCheckIn(ctx, userID, siteID, now.Add(-3*time.Hour), true)
		insertCheckIn(ctx, userID, siteID, now.Add(-2*time.Hour), false) // 期限切れ
		insertNotificationLog(ctx, userID, siteID, now.Add(-30*time.Minute))

		sender := notification.NewFakeEmailSender()
		uc := notify_overdue_sites.NewUsecase(db, sender, slog.Default())

		Expect(uc.Execute(ctx)).To(Succeed())
		Expect(sender.Sent).To(BeEmpty())
	})
})
