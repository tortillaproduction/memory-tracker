package http

import (
	"log/slog"
	"net/http"

	"github.com/tortillaproduction/study-tracker/internal/infrastructure/auth"
	"github.com/tortillaproduction/study-tracker/internal/interface/http/handler"
	"github.com/tortillaproduction/study-tracker/internal/interface/http/middleware"
	"github.com/tortillaproduction/study-tracker/internal/usecase/checkin_site"
	"github.com/tortillaproduction/study-tracker/internal/usecase/register_site"
)

type Dependencies struct {
	RegisterSiteUsecase *register_site.Usecase
	CheckinSiteUsecase  *checkin_site.Usecase
	AuthHandler         *handler.AuthHandler
	SessionStore        *auth.SessionStore
	FrontendURL         string
	Logger              *slog.Logger
}

func NewRouter(deps Dependencies) http.Handler {
	mux := http.NewServeMux()

	siteHandler := handler.NewSiteHandler(deps.RegisterSiteUsecase, deps.Logger)
	checkinHandler := handler.NewCheckInHandler(deps.CheckinSiteUsecase, deps.Logger)
	requireAuth := middleware.RequireAuth(deps.SessionStore)

	// --- 認証不要 ---
	mux.HandleFunc("GET /api/auth/google/login", deps.AuthHandler.LoginRedirect)
	mux.HandleFunc("GET /api/auth/google/callback", deps.AuthHandler.Callback)
	mux.HandleFunc("POST /api/auth/logout", deps.AuthHandler.Logout)

	// --- 認証必須 ---
	mux.Handle("GET /api/auth/me", requireAuth(http.HandlerFunc(deps.AuthHandler.Me)))
	mux.Handle("POST /api/sites", requireAuth(http.HandlerFunc(siteHandler.Register)))
	mux.Handle("GET /go/{siteId}", requireAuth(http.HandlerFunc(checkinHandler.CheckInAndRedirect)))

	// TODO: GET /api/sites（一覧取得）、DELETE /api/sites/{id}、GET /api/streaks 等を追加

	return withCORS(deps.FrontendURL, mux)
}

func withCORS(frontendURL string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", frontendURL)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
