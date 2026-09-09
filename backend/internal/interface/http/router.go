package http

import (
	"log/slog"
	"net/http"

	"github.com/tortillaproduction/memory-tracker/internal/infrastructure/auth"
	"github.com/tortillaproduction/memory-tracker/internal/interface/http/handler"
	"github.com/tortillaproduction/memory-tracker/internal/interface/http/middleware"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/checkin_site"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/delete_site"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/list_sites"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/register_site"
)

type Dependencies struct {
	RegisterSiteUsecase *register_site.Usecase
	ListSiteUsecase     *list_sites.Usecase
	DeleteSiteUsecase   *delete_site.Usecase
	CheckinSiteUsecase  *checkin_site.Usecase
	AuthHandler         *handler.AuthHandler
	SessionStore        *auth.SessionStore
	FrontendURL         string
	Logger              *slog.Logger
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
	mux.HandleFunc("GET /api/auth/google/login", deps.AuthHandler.LoginRedirect)
	mux.HandleFunc("GET /api/auth/google/callback", deps.AuthHandler.Callback)
	mux.HandleFunc("POST /api/auth/logout", deps.AuthHandler.Logout)

	// --- 認証必須 ---
	mux.Handle("GET /api/auth/me", requireAuth(http.HandlerFunc(deps.AuthHandler.Me)))
	mux.Handle("GET /api/sites", requireAuth(http.HandlerFunc(siteHandler.List)))
	mux.Handle("POST /api/sites", requireAuth(http.HandlerFunc(siteHandler.Register)))
	mux.Handle("DELETE /api/sites/{siteId}", requireAuth(http.HandlerFunc(siteHandler.Delete)))
	mux.Handle("GET /go/{siteId}", requireAuth(http.HandlerFunc(checkinHandler.CheckInAndRedirect)))

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
