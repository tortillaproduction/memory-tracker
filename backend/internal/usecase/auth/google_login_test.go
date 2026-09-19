package auth_test

// DBを使わないインメモリのフェイクリポジトリのみで完結するユニットテスト。

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	domainNotification "github.com/tortillaproduction/memory-tracker/internal/domain/notification"
	domainUser "github.com/tortillaproduction/memory-tracker/internal/domain/user"
	pg "github.com/tortillaproduction/memory-tracker/internal/infrastructure/persistence/postgres"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/auth"
)

// fakeUserRepository はGoogleIDをキーにしたインメモリのuser.Repository実装。
type fakeUserRepository struct {
	byGoogleID map[string]*domainUser.User
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{byGoogleID: map[string]*domainUser.User{}}
}

func (r *fakeUserRepository) Save(_ context.Context, u *domainUser.User) (*domainUser.User, bool, error) {
	_, existed := r.byGoogleID[u.GoogleID()]
	r.byGoogleID[u.GoogleID()] = u
	return u, !existed, nil
}

func (r *fakeUserRepository) FindByID(_ context.Context, id domainUser.ID) (*domainUser.User, error) {
	for _, u := range r.byGoogleID {
		if u.ID() == id {
			return u, nil
		}
	}
	return nil, domainUser.ErrNotFound
}

func (r *fakeUserRepository) FindByGoogleID(_ context.Context, googleID string) (*domainUser.User, error) {
	u, ok := r.byGoogleID[googleID]
	if !ok {
		return nil, domainUser.ErrNotFound
	}
	return u, nil
}

// fakeSettingRepository はuser.IDをキーにしたインメモリのnotification.SettingRepository実装。
type fakeSettingRepository struct {
	byUserID map[domainUser.ID]*domainNotification.Setting
}

func newFakeSettingRepository() *fakeSettingRepository {
	return &fakeSettingRepository{byUserID: map[domainUser.ID]*domainNotification.Setting{}}
}

func (r *fakeSettingRepository) Create(_ context.Context, s *domainNotification.Setting) error {
	r.byUserID[s.UserID()] = s
	return nil
}

func (r *fakeSettingRepository) FindByUserID(_ context.Context, userID domainUser.ID) (*domainNotification.Setting, error) {
	s, ok := r.byUserID[userID]
	if !ok {
		return nil, domainNotification.ErrSettingNotFound
	}
	return s, nil
}

func (r *fakeSettingRepository) Update(_ context.Context, s *domainNotification.Setting) error {
	r.byUserID[s.UserID()] = s
	return nil
}

var _ = Describe("Googleログイン", func() {
	It("再ログイン時、保存済みのname/pictureUrlがGoogle側の最新値と異なれば同期される", func() {
		userRepo := newFakeUserRepository()
		uc := auth.NewUsecase(userRepo, newFakeSettingRepository(), pg.NewULIDGenerator())
		ctx := context.Background()

		created, err := uc.Execute(ctx, auth.GoogleUserInfo{
			GoogleID: "google-1",
			Email:    "user@example.com",
			Name:     "Taro Yamada",
			Picture:  "https://example.com/old.png",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(created.Name()).To(Equal("Taro Yamada"))

		// DBの値が何らかの理由で壊れてしまっているケースを再現する
		// （users.nameがサイトURLのような値になってしまっていたケースを想定）。
		created.UpdateName("https://broken-site.example.com")
		_, _, err = userRepo.Save(ctx, created)
		Expect(err).NotTo(HaveOccurred())

		updated, err := uc.Execute(ctx, auth.GoogleUserInfo{
			GoogleID: "google-1",
			Email:    "user@example.com",
			Name:     "Taro Yamada",
			Picture:  "https://example.com/new.png",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(updated.Name()).To(Equal("Taro Yamada"))
		Expect(updated.PictureURL()).To(Equal("https://example.com/new.png"))
	})

	It("Google側の値が変わっていなければ、再保存を行わずそのまま返す", func() {
		userRepo := newFakeUserRepository()
		uc := auth.NewUsecase(userRepo, newFakeSettingRepository(), pg.NewULIDGenerator())
		ctx := context.Background()

		created, err := uc.Execute(ctx, auth.GoogleUserInfo{
			GoogleID: "google-2",
			Email:    "user2@example.com",
			Name:     "Hanako Suzuki",
			Picture:  "https://example.com/pic.png",
		})
		Expect(err).NotTo(HaveOccurred())

		again, err := uc.Execute(ctx, auth.GoogleUserInfo{
			GoogleID: "google-2",
			Email:    "user2@example.com",
			Name:     "Hanako Suzuki",
			Picture:  "https://example.com/pic.png",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(again.ID()).To(Equal(created.ID()))
		Expect(again.Name()).To(Equal("Hanako Suzuki"))
	})
})
