package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/tortillaproduction/memory-tracker/internal/domain/site"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/delete_site"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/list_sites"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/register_site"
)

type SiteHandler struct {
	registerSiteUC *register_site.Usecase
	listSitesUC    *list_sites.Usecase
	deleteSiteUC   *delete_site.Usecase
	logger         *slog.Logger
}

func NewSiteHandler(
	registerUC *register_site.Usecase,
	listUC *list_sites.Usecase,
	deleteUC *delete_site.Usecase,
	logger *slog.Logger,
) *SiteHandler {
	return &SiteHandler{
		registerSiteUC: registerUC,
		listSitesUC:    listUC,
		deleteSiteUC:   deleteUC,
		logger:         logger,
	}
}

// List は GET /api/sites のハンドラー。
// サイト一覧とユーザー全体のストリークをまとめて返す。
func (h *SiteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	sites, userStreak, err := h.listSitesUC.Execute(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to list sites", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type siteResponse struct {
		ID                  string  `json:"id"`
		Name                string  `json:"name"`
		URL                 string  `json:"url"`
		IntervalHours       int     `json:"intervalHours"`
		HoursSinceLastCheck float64 `json:"hoursSinceLastCheck"`
		IsOverdue           bool    `json:"isOverdue"`
		SiteStreak          int     `json:"siteStreak"`
	}

	siteList := make([]siteResponse, 0, len(sites))
	for _, s := range sites {
		siteList = append(siteList, siteResponse{
			ID:                  string(s.ID),
			Name:                s.Name,
			URL:                 s.URL,
			IntervalHours:       s.IntervalHours,
			HoursSinceLastCheck: s.HoursSinceLastCheck,
			IsOverdue:           s.IsOverdue,
			SiteStreak:          s.SiteStreak,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"sites":      siteList,
		"userStreak": userStreak,
	})
}

// Register は POST /api/sites のハンドラー。
func (h *SiteHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name          string `json:"name"`
		URL           string `json:"url"`
		IntervalHours int    `json:"intervalHours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.URL == "" {
		http.Error(w, "name and url are required", http.StatusBadRequest)
		return
	}

	userID := userIDFromContext(r.Context()) // TODO: auth middlewareがセットする想定

	s, err := h.registerSiteUC.Execute(r.Context(), register_site.Input{
		UserID:        userID,
		Name:          req.Name,
		URL:           req.URL,
		IntervalHours: req.IntervalHours,
	})
	if err != nil {
		if errors.Is(err, register_site.ErrSiteLimitReached) {
			http.Error(w, "site limit reached for your plan", http.StatusPaymentRequired)
			return
		}
		h.logger.Error("failed to register site", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": string(s.ID())})
}

// Delete は DELETE /api/sites/{siteId} のハンドラー。
func (h *SiteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	siteID := site.ID(r.PathValue("siteId"))
	userID := userIDFromContext(r.Context())

	if err := h.deleteSiteUC.Execute(r.Context(), userID, siteID); err != nil {
		if errors.Is(err, delete_site.ErrNotFound) {
			http.Error(w, "site not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, delete_site.ErrForbidden) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		h.logger.Error("failed to delete site", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
