package get_gate_candidates_test

// 実DB(Postgres)を使う統合テスト。DBに接続できない場合はスキップする。

import (
	"context"
	"database/sql"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tortillaproduction/memory-tracker/internal/domain/gate"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
	pg "github.com/tortillaproduction/memory-tracker/internal/infrastructure/persistence/postgres"
	"github.com/tortillaproduction/memory-tracker/internal/testutil/dbtest"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/dismiss_gate"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/get_gate_candidates"
)

var (
	db          *sql.DB
	dbAvailable bool
)

var _ = BeforeSuite(func() {
	db, dbAvailable = dbtest.Open()
})

var _ = AfterSuite(func() {
	if db != nil {
		db.Close()
	}
})

// now は日本時間 2026-01-15 00:30。日付の境界(0時)の前後を検証しやすいよう固定する。
var now = time.Date(2026, 1, 15, 0, 30, 0, 0, gate.Location)

var _ = Describe("ゲートの候補取得", func() {
	var (
		ctx    context.Context
		userID string
		others []string
		uc     *get_gate_candidates.Usecase
	)

	BeforeEach(func() {
		if !dbAvailable {
			Skip("DATABASE_URLに接続できないためスキップ (docker compose up -d db を実行してください)")
		}
		ctx = context.Background()
		userID = dbtest.CreateUser(ctx, db)
		uc = get_gate_candidates.NewUsecase(
			pg.NewOverdueSiteQuery(db),
			pg.NewCheckInRepository(db),
			pg.NewGateDismissalRepository(db),
		)
	})

	AfterEach(func() {
		for _, id := range append(others, userID) {
			if id != "" {
				dbtest.CleanupUser(ctx, db, id)
			}
		}
		userID, others = "", nil
	})

	// overdueSite は、nowの時点で期限をoverdueBy過ぎたサイトを作る(interval 24時間、前回の訪問は前日より前)。
	overdueSite := func(name string, overdueBy time.Duration) string {
		siteID := dbtest.CreateSite(ctx, db, userID, name, 24)
		last := now.Add(-24*time.Hour - overdueBy)
		dbtest.InsertCheckIn(ctx, db, userID, siteID, last.Add(-48*time.Hour), true)
		dbtest.InsertCheckIn(ctx, db, userID, siteID, last, false)
		return siteID
	}

	It("期限切れサイトを超過時間の長い順に最大3件返す", func() {
		overdueSite("2h", 2*time.Hour)
		overdueSite("30h", 30*time.Hour)
		overdueSite("1h", 1*time.Hour)
		overdueSite("10h", 10*time.Hour)

		res, err := uc.Execute(ctx, user.ID(userID), now)
		Expect(err).NotTo(HaveOccurred())

		Expect(res.AlreadyDoneToday).To(BeFalse())
		Expect(res.Candidates).To(HaveLen(3))
		Expect([]string{res.Candidates[0].Name, res.Candidates[1].Name, res.Candidates[2].Name}).
			To(Equal([]string{"30h", "10h", "2h"}))
		Expect(res.Candidates[0].OverdueBy).To(Equal(30 * time.Hour))
		Expect(res.Candidates[0].URL).To(HavePrefix("https://example.com/"))
	})

	It("期限内・アーカイブ済み・他のユーザーのサイトは候補に含めない", func() {
		fresh := dbtest.CreateSite(ctx, db, userID, "fresh", 24)
		dbtest.InsertCheckIn(ctx, db, userID, fresh, now.Add(-48*time.Hour), true)
		dbtest.InsertCheckIn(ctx, db, userID, fresh, now.Add(-2*time.Hour), false)

		archived := overdueSite("archived", 5*time.Hour)
		dbtest.ArchiveSite(ctx, db, archived)

		other := dbtest.CreateUser(ctx, db)
		others = append(others, other)
		otherSite := dbtest.CreateSite(ctx, db, other, "other", 1)
		dbtest.InsertCheckIn(ctx, db, other, otherSite, now.Add(-72*time.Hour), true)

		res, err := uc.Execute(ctx, user.ID(userID), now)
		Expect(err).NotTo(HaveOccurred())
		Expect(res.Candidates).To(BeEmpty())
		Expect(res.AlreadyDoneToday).To(BeFalse())
	})

	It("通知済み(notification_logsあり)や一度も訪問していない期限切れサイトも候補に含める", func() {
		notified := overdueSite("notified", 3*time.Hour)
		dbtest.InsertNotificationLog(ctx, db, userID, notified, now.Add(-10*time.Minute))

		neverVisited := dbtest.CreateSite(ctx, db, userID, "never", 24)
		dbtest.InsertCheckIn(ctx, db, userID, neverVisited, now.Add(-30*time.Hour), true)

		res, err := uc.Execute(ctx, user.ID(userID), now)
		Expect(err).NotTo(HaveOccurred())
		Expect(res.Candidates).To(HaveLen(2))
		Expect(res.Candidates[0].Name).To(Equal("never")) // 6時間超過 > 3時間超過
		Expect(res.Candidates[1].Name).To(Equal("notified"))
	})

	Describe("今日済みの判定(日本時間)", func() {
		var siteID string

		BeforeEach(func() {
			siteID = overdueSite("overdue", 5*time.Hour)
		})

		It("今日0時以降にどれかのサイトへチェックインしていれば、今日済みとして空の候補を返す", func() {
			fresh := dbtest.CreateSite(ctx, db, userID, "fresh", 24)
			dbtest.InsertCheckIn(ctx, db, userID, fresh, now.Add(-72*time.Hour), true)
			dbtest.InsertCheckIn(ctx, db, userID, fresh, gate.StartOfDay(now).Add(10*time.Minute), false)

			res, err := uc.Execute(ctx, user.ID(userID), now)
			Expect(err).NotTo(HaveOccurred())
			Expect(res.AlreadyDoneToday).To(BeTrue())
			Expect(res.Candidates).NotTo(BeNil())
			Expect(res.Candidates).To(BeEmpty())
		})

		It("前日23:50のチェックインは今日に数えない", func() {
			fresh := dbtest.CreateSite(ctx, db, userID, "fresh", 24)
			dbtest.InsertCheckIn(ctx, db, userID, fresh, now.Add(-72*time.Hour), true)
			dbtest.InsertCheckIn(ctx, db, userID, fresh, gate.StartOfDay(now).Add(-10*time.Minute), false)

			res, err := uc.Execute(ctx, user.ID(userID), now)
			Expect(err).NotTo(HaveOccurred())
			Expect(res.AlreadyDoneToday).To(BeFalse())
			Expect(res.Candidates).To(HaveLen(1))
			Expect(res.Candidates[0].SiteID).To(Equal(siteID))
		})

		It("登録時の初回チェックイン(is_initial)は今日に数えない", func() {
			registered := dbtest.CreateSite(ctx, db, userID, "registered today", 24)
			dbtest.InsertCheckIn(ctx, db, userID, registered, gate.StartOfDay(now).Add(5*time.Minute), true)

			res, err := uc.Execute(ctx, user.ID(userID), now)
			Expect(err).NotTo(HaveOccurred())
			Expect(res.AlreadyDoneToday).To(BeFalse())
		})

		It("今日脱出口で解除していれば今日済み、前日の解除は数えない", func() {
			dismiss := dismiss_gate.NewUsecase(pg.NewGateDismissalRepository(db), pg.NewULIDGenerator())

			Expect(dismiss.Execute(ctx, user.ID(userID), now.Add(-time.Hour))).To(Succeed()) // 前日23:30
			res, err := uc.Execute(ctx, user.ID(userID), now)
			Expect(err).NotTo(HaveOccurred())
			Expect(res.AlreadyDoneToday).To(BeFalse())

			Expect(dismiss.Execute(ctx, user.ID(userID), now)).To(Succeed())
			Expect(dismiss.Execute(ctx, user.ID(userID), now)).To(Succeed()) // 同じ日の2回目は冪等
			res, err = uc.Execute(ctx, user.ID(userID), now)
			Expect(err).NotTo(HaveOccurred())
			Expect(res.AlreadyDoneToday).To(BeTrue())
		})
	})
})
