package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/tortillaproduction/memory-tracker/internal/domain/site"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
	"github.com/tortillaproduction/memory-tracker/internal/interface/http/middleware"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/checkin_site"
)

type CheckInHandler struct {
	checkinSiteUC *checkin_site.Usecase
	logger        *slog.Logger
}

func NewCheckInHandler(uc *checkin_site.Usecase, logger *slog.Logger) *CheckInHandler {
	return &CheckInHandler{checkinSiteUC: uc, logger: logger}
}

// CheckInAndRedirect は /go/{siteId} を踏んだときに呼ばれる。
// チェックインを記録した上で、実際のサイトへ302リダイレクトする。
func (h *CheckInHandler) CheckInAndRedirect(w http.ResponseWriter, r *http.Request) {
	siteID := site.ID(r.PathValue("siteId"))
	userID := userIDFromContext(r.Context())

	redirectURL, err := h.checkinSiteUC.Execute(r.Context(), userID, siteID)
	if err != nil {
		h.logger.Error("failed to checkin", "error", err, "siteId", siteID)
		http.Error(w, "site not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// userIDFromContext は認証ミドルウェアがセットしたユーザーIDを取り出す想定のヘルパー。
func userIDFromContext(ctx context.Context) user.ID {
	if v, ok := ctx.Value(middleware.UserIDContextKey).(user.ID); ok {
		return v
	}
	return ""
}
