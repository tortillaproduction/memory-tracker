package checkin_site_test

// 実DB(Postgres)を使う統合テスト。DBに接続できない場合はスキップする。

import (
	"context"
	"database/sql"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tortillaproduction/memory-tracker/internal/domain/site"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
	pg "github.com/tortillaproduction/memory-tracker/internal/infrastructure/persistence/postgres"
	"github.com/tortillaproduction/memory-tracker/internal/testutil/dbtest"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/checkin_site"
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

var _ = Describe("チェックイン", func() {
	var (
		ctx    context.Context
		userID string
		siteID string
		other  string
		uc     *checkin_site.Usecase
	)

	BeforeEach(func() {
		if !dbAvailable {
			Skip("DATABASE_URLに接続できないためスキップ (docker compose up -d db を実行してください)")
		}
		ctx = context.Background()
		userID = dbtest.CreateUser(ctx, db)
		siteID = dbtest.CreateSite(ctx, db, userID, "site", 24)
		uc = checkin_site.NewUsecase(pg.NewSiteRepository(db), pg.NewCheckInRepository(db), pg.NewULIDGenerator())
	})

	AfterEach(func() {
		dbtest.CleanupUser(ctx, db, userID)
		if other != "" {
			dbtest.CleanupUser(ctx, db, other)
			other = ""
		}
	})

	It("チェックインを記録し、サイトのURLを返す", func() {
		url, err := uc.Execute(ctx, user.ID(userID), site.ID(siteID))
		Expect(err).NotTo(HaveOccurred())
		Expect(url).To(Equal("https://example.com/" + siteID))
		Expect(dbtest.CountCheckIns(ctx, db, siteID)).To(Equal(1))
	})

	It("5分以内に同じサイトへ再度チェックインしても記録は増えないが、URLは返す", func() {
		_, err := uc.Execute(ctx, user.ID(userID), site.ID(siteID))
		Expect(err).NotTo(HaveOccurred())

		url, err := uc.Execute(ctx, user.ID(userID), site.ID(siteID))
		Expect(err).NotTo(HaveOccurred())
		Expect(url).To(Equal("https://example.com/" + siteID))
		Expect(dbtest.CountCheckIns(ctx, db, siteID)).To(Equal(1))
	})

	It("前回のチェックインから5分を過ぎていれば記録する", func() {
		dbtest.InsertCheckIn(ctx, db, userID, siteID, time.Now().Add(-6*time.Minute), false)

		_, err := uc.Execute(ctx, user.ID(userID), site.ID(siteID))
		Expect(err).NotTo(HaveOccurred())
		Expect(dbtest.CountCheckIns(ctx, db, siteID)).To(Equal(2))
	})

	It("登録時の初回チェックイン(is_initial)は重複判定に使わない", func() {
		dbtest.InsertCheckIn(ctx, db, userID, siteID, time.Now(), true)

		_, err := uc.Execute(ctx, user.ID(userID), site.ID(siteID))
		Expect(err).NotTo(HaveOccurred())
		Expect(dbtest.CountCheckIns(ctx, db, siteID)).To(Equal(1))
	})

	It("他のユーザーのサイトにはチェックインできない", func() {
		other = dbtest.CreateUser(ctx, db)

		_, err := uc.Execute(ctx, user.ID(other), site.ID(siteID))
		Expect(err).To(MatchError(checkin_site.ErrSiteNotFound))
		Expect(dbtest.CountCheckIns(ctx, db, siteID)).To(Equal(0))
	})

	It("アーカイブ済み・存在しないサイトにはチェックインできない", func() {
		dbtest.ArchiveSite(ctx, db, siteID)
		_, err := uc.Execute(ctx, user.ID(userID), site.ID(siteID))
		Expect(err).To(MatchError(checkin_site.ErrSiteNotFound))

		_, err = uc.Execute(ctx, user.ID(userID), site.ID("no-such-site"))
		Expect(err).To(MatchError(checkin_site.ErrSiteNotFound))
	})
})
