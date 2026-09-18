package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/tortillaproduction/memory-tracker/internal/domain/site"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
	"github.com/tortillaproduction/memory-tracker/internal/infrastructure/auth"
)

type contextKey string

const UserIDContextKey contextKey = "userID"

const SeesionCookieName = "session_id"

// RequireAuth はCookieのセッションIDを検証し、有効ならユーザーIDをcontextに詰めて次に渡す。
// 無効・未ログインの場合は401を返す。
func RequireAuth(sessionStore *auth.SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := authenticateByCookie(w, r, sessionStore)
			if !ok {
				return
			}
			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CheckinTokenVerifier はメールの/go/{siteId}リンクに埋め込まれたトークンを検証する。
type CheckinTokenVerifier interface {
	VerifyCheckinToken(siteID site.ID, token string) (user.ID, error)
}

// RequireAuthOrCheckinToken は /go/{siteId} 専用の認証ミドルウェア。
// クエリに?token=があればそれをセッションCookieの代わりとして検証する
// (スマホのメールアプリはアプリ内WebViewで開くことが多く、ログイン中の
// ブラウザとCookieが共有されないため)。tokenが無ければ通常のCookie認証にフォールバックする。
func RequireAuthOrCheckinToken(sessionStore *auth.SessionStore, verifier CheckinTokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token := r.URL.Query().Get("token"); token != "" {
				userID, err := verifier.VerifyCheckinToken(site.ID(r.PathValue("siteId")), token)
				if err != nil {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			userID, ok := authenticateByCookie(w, r, sessionStore)
			if !ok {
				return
			}
			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// authenticateByCookie はセッションCookieを検証する。失敗時は自身でエラーレスポンスを
// 書き込み、falseを返す(呼び出し側はそのままハンドラーを終了させること)。
func authenticateByCookie(w http.ResponseWriter, r *http.Request, sessionStore *auth.SessionStore) (user.ID, bool) {
	cookie, err := r.Cookie(SeesionCookieName)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return "", false
	}

	userID, err := sessionStore.FindUserID(r.Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, auth.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return "", false
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return "", false
	}

	return userID, true
}
