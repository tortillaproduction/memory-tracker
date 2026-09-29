package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/authenticate_gate_token"
)

// GateTokenAuthenticator はゲート(Androidアプリ)の認証トークンを検証する。
// 無効なトークンには authenticate_gate_token.ErrInvalidToken を返す。
type GateTokenAuthenticator interface {
	Execute(ctx context.Context, plain string, now time.Time) (user.ID, error)
}

// RequireGateToken は Authorization: Bearer <token> でユーザーを特定する、ゲートAPI専用の認証ミドルウェア。
// トークンは秘匿情報なので、失敗時もログやレスポンスに含めない。
func RequireGateToken(authenticator GateTokenAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			plain, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || plain == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			userID, err := authenticator.Execute(r.Context(), plain, time.Now())
			if errors.Is(err, authenticate_gate_token.ErrInvalidToken) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
