package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/tortillaproduction/memory-tracker/internal/usecase/register_site"
)

type SiteHandler struct {
	registerSiteUC *register_site.Usecase
	logger         *slog.Logger
}

func NewSiteHandler(uc *register_site.Usecase, logger *slog.Logger) *SiteHandler {
	return &SiteHandler{registerSiteUC: uc, logger: logger}
}

type registerSiteRequest struct {
	Name          string `json:"name"`
	URL           string `json:"url"`
	IntervalHours int    `json:"interval_hours"`
}

func (h *SiteHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerSiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userID := userIDFromContext(r.Context()) // TODO: auth middlewareがセットする想定

	site, err := h.registerSiteUC.Execute(r.Context(), register_site.Input{
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
	json.NewEncoder(w).Encode(map[string]string{"id": string(site.ID())})
}
