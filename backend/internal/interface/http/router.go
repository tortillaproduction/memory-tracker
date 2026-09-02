package http

import (
	"log/slog"
	"net/http"

	"github.com/tortillaproduction/study-tracker/internal/interface/http/handler"
	"github.com/tortillaproduction/study-tracker/internal/usecase/checkin_site"
	"github.com/tortillaproduction/study-tracker/internal/usecase/register_site"
)

type Dependencies struct {
	RegisterSiteUsecase *register_site.Usecase
	CheckinSiteUsecase  *checkin_site.Usecase
	Logger              *slog.Logger
}

func NewRouter(deps Dependencies) http.Handler {
	mux := http.NewServeMux()

	siteHandler := handler.NewSiteHandler(deps.RegisterSiteUsecase, deps.Logger)
	checkinHandler := handler.NewCheckInHandler(deps.CheckinSiteUsecase, deps.Logger)

	mux.HandleFunc("POST /api/sites", siteHandler.Register)
	mux.HandleFunc("GET /go/{siteId}", checkinHandler.CheckInAndRedirect) // 経由リンク方式

	// TODO: /api/auth/google, /api/sites/{id}, /api/streaks 等を追加

	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
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
