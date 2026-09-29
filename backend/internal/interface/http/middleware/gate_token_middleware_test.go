package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
	"github.com/tortillaproduction/memory-tracker/internal/interface/http/middleware"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/authenticate_gate_token"
)

// fakeAuthenticator は "valid-token" だけを受け付ける。err が設定されていればそれを返す。
type fakeAuthenticator struct {
	err error
}

func (a *fakeAuthenticator) Execute(_ context.Context, plain string, _ time.Time) (user.ID, error) {
	if a.err != nil {
		return "", a.err
	}
	if plain != "valid-token" {
		return "", authenticate_gate_token.ErrInvalidToken
	}
	return user.ID("user-1"), nil
}

var _ = Describe("RequireGateToken", func() {
	var (
		authenticator *fakeAuthenticator
		gotUserID     user.ID
		called        bool
	)

	serve := func(authorization string) int {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			gotUserID, _ = r.Context().Value(middleware.UserIDContextKey).(user.ID)
			w.WriteHeader(http.StatusOK)
		})
		req := httptest.NewRequest(http.MethodGet, "/api/gate/candidates", nil)
		if authorization != "" {
			req.Header.Set("Authorization", authorization)
		}
		rec := httptest.NewRecorder()
		middleware.RequireGateToken(authenticator)(next).ServeHTTP(rec, req)
		return rec.Code
	}

	BeforeEach(func() {
		authenticator = &fakeAuthenticator{}
		gotUserID, called = "", false
	})

	It("正しいBearerトークンならユーザーIDをcontextに入れて次に渡す", func() {
		Expect(serve("Bearer valid-token")).To(Equal(http.StatusOK))
		Expect(called).To(BeTrue())
		Expect(gotUserID).To(Equal(user.ID("user-1")))
	})

	DescribeTable("トークンが無い・形式が違う・無効なら401を返し、次に渡さない",
		func(authorization string) {
			Expect(serve(authorization)).To(Equal(http.StatusUnauthorized))
			Expect(called).To(BeFalse())
		},
		Entry("ヘッダーなし", ""),
		Entry("Bearerの後が空", "Bearer "),
		Entry("Bearer以外の方式", "Basic valid-token"),
		Entry("無効なトークン", "Bearer wrong-token"),
	)

	It("照合中にDBエラーなどが起きた場合は500を返す", func() {
		authenticator.err = errors.New("db down")
		Expect(serve("Bearer valid-token")).To(Equal(http.StatusInternalServerError))
		Expect(called).To(BeFalse())
	})
})
