package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/tortillaproduction/study-tracker/internal/infrastructure/auth"
)

type contextKey string

const UserIDContextKey contextKey = "userID"

const SeesionCookieName = "session_id"

// RequireAuth はCookieのセッションIDを検証し、有効ならユーザーIDをcontextに詰めて次に渡す。
// 無効・未ログインの場合は401を返す。
func RequireAuth(sessionStore *auth.SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SeesionCookieName)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := sessionStore.FindUserID(r.Context(), cookie.Value)
			if err != nil {
				if errors.Is(err, auth.ErrSessionNotFound) {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
