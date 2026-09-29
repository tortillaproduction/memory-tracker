package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/site"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/checkin_site"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/dismiss_gate"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/get_gate_candidates"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/issue_gate_token"
)

// GateHandler はゲート(Androidアプリ)向けのAPIと、Webからのトークン発行APIを扱う。
// 認証トークンの平文は発行時のレスポンスにだけ含め、ログには出さない。
type GateHandler struct {
	candidatesUC *get_gate_candidates.Usecase
	checkinUC    *checkin_site.Usecase
	dismissUC    *dismiss_gate.Usecase
	issueTokenUC *issue_gate_token.Usecase
	// apiBaseURL はアプリがAPIを呼ぶときのベースURL。QRコードに埋め込む。
	apiBaseURL string
	logger     *slog.Logger
}

func NewGateHandler(
	candidatesUC *get_gate_candidates.Usecase,
	checkinUC *checkin_site.Usecase,
	dismissUC *dismiss_gate.Usecase,
	issueTokenUC *issue_gate_token.Usecase,
	apiBaseURL string,
	logger *slog.Logger,
) *GateHandler {
	return &GateHandler{
		candidatesUC: candidatesUC,
		checkinUC:    checkinUC,
		dismissUC:    dismissUC,
		issueTokenUC: issueTokenUC,
		apiBaseURL:   apiBaseURL,
		logger:       logger,
	}
}

type gateCandidateResponse struct {
	SiteID       string  `json:"siteId"`
	Name         string  `json:"name"`
	URL          string  `json:"url"`
	OverdueHours float64 `json:"overdueHours"`
}

// Candidates は GET /api/gate/candidates のハンドラー(Bearer認証)。
func (h *GateHandler) Candidates(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	res, err := h.candidatesUC.Execute(r.Context(), userID, time.Now())
	if err != nil {
		h.logger.Error("failed to get gate candidates", "userID", userID, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	candidates := make([]gateCandidateResponse, 0, len(res.Candidates))
	for _, c := range res.Candidates {
		candidates = append(candidates, gateCandidateResponse{
			SiteID:       c.SiteID,
			Name:         c.Name,
			URL:          c.URL,
			OverdueHours: math.Round(c.OverdueBy.Hours()*10) / 10,
		})
	}

	writeJSON(w, map[string]any{
		"alreadyDoneToday": res.AlreadyDoneToday,
		"candidates":       candidates,
	})
}

// Checkin は POST /api/gate/checkin のハンドラー(Bearer認証)。
// チェックインを記録し、アプリがブラウザで開くサイトのURLを返す。
func (h *GateHandler) Checkin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SiteID string `json:"siteId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SiteID == "" {
		http.Error(w, "siteId is required", http.StatusBadRequest)
		return
	}

	userID := userIDFromContext(r.Context())

	url, err := h.checkinUC.Execute(r.Context(), userID, site.ID(req.SiteID))
	if errors.Is(err, checkin_site.ErrSiteNotFound) {
		http.Error(w, "site not found", http.StatusNotFound)
		return
	}
	if err != nil {
		h.logger.Error("failed to checkin from gate", "userID", userID, "siteId", req.SiteID, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"url": url})
}

// Dismiss は POST /api/gate/dismiss のハンドラー(Bearer認証)。
// 脱出口でゲートを解除したことを記録し、その日はゲートを出さないようにする。
func (h *GateHandler) Dismiss(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	if err := h.dismissUC.Execute(r.Context(), userID, time.Now()); err != nil {
		h.logger.Error("failed to record gate dismissal", "userID", userID, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// IssueToken は POST /api/gate/token のハンドラー(Cookie認証)。
// トークンを発行(再発行)し、平文とQRコードに埋め込む文字列を返す。平文を取得できるのはこの時だけ。
func (h *GateHandler) IssueToken(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	plain, err := h.issueTokenUC.Execute(r.Context(), userID, time.Now())
	if err != nil {
		h.logger.Error("failed to issue gate token", "userID", userID, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	qrPayload, err := json.Marshal(map[string]any{
		"v":       1,
		"baseUrl": h.apiBaseURL,
		"token":   plain,
	})
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, map[string]string{
		"token":     plain,
		"qrPayload": string(qrPayload),
	})
}

// TokenStatus は GET /api/gate/token のハンドラー(Cookie認証)。平文は返さない。
func (h *GateHandler) TokenStatus(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	st, err := h.issueTokenUC.Status(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get gate token status", "userID", userID, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"exists":     st.Exists,
		"createdAt":  st.CreatedAt,
		"lastUsedAt": st.LastUsedAt,
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
