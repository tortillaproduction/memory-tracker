package http

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/tortillaproduction/memory-tracker/internal/infrastructure/auth"
	"github.com/tortillaproduction/memory-tracker/internal/interface/http/handler"
	"github.com/tortillaproduction/memory-tracker/internal/interface/http/middleware"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/checkin_site"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/delete_site"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/list_sites"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/register_site"
)

type Dependencies struct {
	RegisterSiteUsecase  *register_site.Usecase
	ListSiteUsecase      *list_sites.Usecase
	DeleteSiteUsecase    *delete_site.Usecase
	CheckinSiteUsecase   *checkin_site.Usecase
	AuthHandler          *handler.AuthHandler
	PushHandler          *handler.PushHandler
	SessionStore         *auth.SessionStore
	CheckinTokenVerifier middleware.CheckinTokenVerifier
	FrontendURL          string
	Logger               *slog.Logger
}

func NewRouter(deps Dependencies) http.Handler {
	mux := http.NewServeMux()

	siteHandler := handler.NewSiteHandler(
		deps.RegisterSiteUsecase,
		deps.ListSiteUsecase,
		deps.DeleteSiteUsecase,
		deps.Logger,
	)
	checkinHandler := handler.NewCheckInHandler(deps.CheckinSiteUsecase, deps.Logger)
	requireAuth := middleware.RequireAuth(deps.SessionStore)

	// --- 認証不要 ---
	// keep-alive用。DBには触れず、インスタンスを起こすためだけに使う。
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /api/auth/google/login", deps.AuthHandler.LoginRedirect)
	mux.HandleFunc("GET /api/auth/google/callback", deps.AuthHandler.Callback)
	mux.HandleFunc("POST /api/auth/logout", deps.AuthHandler.Logout)
	mux.HandleFunc("GET /api/push/vapid-public-key", deps.PushHandler.VAPIDPublicKey)

	// --- 認証必須 ---
	mux.Handle("GET /api/auth/me", requireAuth(http.HandlerFunc(deps.AuthHandler.Me)))
	mux.Handle("GET /api/sites", requireAuth(http.HandlerFunc(siteHandler.List)))
	mux.Handle("POST /api/sites", requireAuth(http.HandlerFunc(siteHandler.Register)))
	mux.Handle("DELETE /api/sites/{siteId}", requireAuth(http.HandlerFunc(siteHandler.Delete)))
	mux.Handle("POST /api/push/subscribe", requireAuth(http.HandlerFunc(deps.PushHandler.Subscribe)))
	mux.Handle("POST /api/push/unsubscribe", requireAuth(http.HandlerFunc(deps.PushHandler.Unsubscribe)))
	mux.Handle("GET /api/notification-preferences", requireAuth(http.HandlerFunc(deps.PushHandler.GetPreferences)))
	mux.Handle("PATCH /api/notification-preferences", requireAuth(http.HandlerFunc(deps.PushHandler.UpdatePreferences)))
	requireAuthOrCheckinToken := middleware.RequireAuthOrCheckinToken(deps.SessionStore, deps.CheckinTokenVerifier)
	mux.Handle("GET /go/{siteId}", requireAuthOrCheckinToken(http.HandlerFunc(checkinHandler.CheckInAndRedirect)))

	return withRecover(deps.Logger, withCORS(deps.FrontendURL, mux))
}

// withRecover はハンドラー内でのpanicを捕捉し、コネクションを切断させる代わりに
// 500を返す。recoverが無いとGoの標準http.Serverはpanic時にコネクションを
// 切断するだけになり、クライアント側には原因不明のネットワークエラーとして
// 見えてしまうため、切り分けを容易にする目的で入れている。
func withRecover(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("panic recovered", "error", rec, "stack", string(debug.Stack()))
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func withCORS(frontendURL string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", frontendURL)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
