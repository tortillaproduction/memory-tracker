package authenticate_gate_token_test

// 実DB(Postgres)を使う統合テスト。トークンの発行(issue_gate_token)と照合をまとめて検証する。

import (
	"context"
	"database/sql"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"

	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
	"github.com/tortillaproduction/memory-tracker/internal/infrastructure/auth"
	pg "github.com/tortillaproduction/memory-tracker/internal/infrastructure/persistence/postgres"
	"github.com/tortillaproduction/memory-tracker/internal/testutil/dbtest"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/authenticate_gate_token"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/issue_gate_token"
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

var _ = Describe("ゲート用トークンの認証", func() {
	var (
		ctx          context.Context
		userID       string
		issue        *issue_gate_token.Usecase
		authenticate *authenticate_gate_token.Usecase
		now          time.Time
	)

	lastUsedAt := func() *time.Time {
		var t sql.NullTime
		Expect(db.QueryRowContext(ctx,
			`SELECT last_used_at FROM gate_tokens WHERE user_id = $1`, userID,
		).Scan(&t)).To(Succeed())
		if !t.Valid {
			return nil
		}
		return &t.Time
	}

	BeforeEach(func() {
		if !dbAvailable {
			Skip("DATABASE_URLに接続できないためスキップ (docker compose up -d db を実行してください)")
		}
		ctx = context.Background()
		userID = dbtest.CreateUser(ctx, db)
		repo := pg.NewGateTokenRepository(db)
		issue = issue_gate_token.NewUsecase(repo, auth.NewGateTokenGenerator(), pg.NewULIDGenerator())
		authenticate = authenticate_gate_token.NewUsecase(repo, auth.HashGateToken)
		now = time.Now().Truncate(time.Second)
	})

	AfterEach(func() {
		dbtest.CleanupUser(ctx, db, userID)
	})

	It("発行したトークンで持ち主のユーザーを特定でき、DBには平文ではなくハッシュが保存される", func() {
		plain, err := issue.Execute(ctx, user.ID(userID), now)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(plain)).To(BeNumerically(">=", 43))

		var stored string
		Expect(db.QueryRowContext(ctx,
			`SELECT token_hash FROM gate_tokens WHERE user_id = $1`, userID,
		).Scan(&stored)).To(Succeed())
		Expect(stored).NotTo(Equal(plain))
		Expect(stored).To(Equal(auth.HashGateToken(plain)))

		got, err := authenticate.Execute(ctx, plain, now)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(user.ID(userID)))
	})

	It("不正なトークン・空のトークンはErrInvalidTokenになる", func() {
		_, err := issue.Execute(ctx, user.ID(userID), now)
		Expect(err).NotTo(HaveOccurred())

		_, err = authenticate.Execute(ctx, "not-a-valid-token", now)
		Expect(err).To(MatchError(authenticate_gate_token.ErrInvalidToken))

		_, err = authenticate.Execute(ctx, "", now)
		Expect(err).To(MatchError(authenticate_gate_token.ErrInvalidToken))
	})

	It("再発行すると以前のトークンは使えなくなる", func() {
		oldPlain, err := issue.Execute(ctx, user.ID(userID), now)
		Expect(err).NotTo(HaveOccurred())
		newPlain, err := issue.Execute(ctx, user.ID(userID), now.Add(time.Minute))
		Expect(err).NotTo(HaveOccurred())

		_, err = authenticate.Execute(ctx, oldPlain, now)
		Expect(err).To(MatchError(authenticate_gate_token.ErrInvalidToken))

		got, err := authenticate.Execute(ctx, newPlain, now)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(user.ID(userID)))

		var count int
		Expect(db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM gate_tokens WHERE user_id = $1`, userID,
		).Scan(&count)).To(Succeed())
		Expect(count).To(Equal(1))
	})

	It("last_used_atは初回と、前回から1時間以上経ったときだけ更新される", func() {
		plain, err := issue.Execute(ctx, user.ID(userID), now)
		Expect(err).NotTo(HaveOccurred())
		Expect(lastUsedAt()).To(BeNil())

		_, err = authenticate.Execute(ctx, plain, now)
		Expect(err).NotTo(HaveOccurred())
		Expect(lastUsedAt()).To(gstruct.PointTo(BeTemporally("==", now)))

		_, err = authenticate.Execute(ctx, plain, now.Add(30*time.Minute))
		Expect(err).NotTo(HaveOccurred())
		Expect(lastUsedAt()).To(gstruct.PointTo(BeTemporally("==", now)))

		_, err = authenticate.Execute(ctx, plain, now.Add(2*time.Hour))
		Expect(err).NotTo(HaveOccurred())
		Expect(lastUsedAt()).To(gstruct.PointTo(BeTemporally("==", now.Add(2*time.Hour))))
	})

	It("トークンの状態には平文を含めず、発行の有無と日時だけを返す", func() {
		st, err := issue.Status(ctx, user.ID(userID))
		Expect(err).NotTo(HaveOccurred())
		Expect(st.Exists).To(BeFalse())

		_, err = issue.Execute(ctx, user.ID(userID), now)
		Expect(err).NotTo(HaveOccurred())

		st, err = issue.Status(ctx, user.ID(userID))
		Expect(err).NotTo(HaveOccurred())
		Expect(st.Exists).To(BeTrue())
		Expect(st.CreatedAt).To(gstruct.PointTo(BeTemporally("==", now)))
		Expect(st.LastUsedAt).To(BeNil())
	})
})
