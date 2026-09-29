package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/tortillaproduction/memory-tracker/internal/domain/notification"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/update_notification_mode"
)

type NotificationSettingsHandler struct {
	modeUC *update_notification_mode.Usecase
	logger *slog.Logger
}

func NewNotificationSettingsHandler(modeUC *update_notification_mode.Usecase, logger *slog.Logger) *NotificationSettingsHandler {
	return &NotificationSettingsHandler{modeUC: modeUC, logger: logger}
}

// Get は GET /api/notification-settings のハンドラー。
func (h *NotificationSettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	mode, err := h.modeUC.Get(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get notification settings", "userID", userID, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"mode": string(mode)})
}

// Update は PATCH /api/notification-settings のハンドラー。通知モード(email / gate)を切り替える。
func (h *NotificationSettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	mode, err := notification.ParseMode(req.Mode)
	if err != nil {
		http.Error(w, "mode must be 'email' or 'gate'", http.StatusBadRequest)
		return
	}

	userID := userIDFromContext(r.Context())

	updated, err := h.modeUC.Update(r.Context(), userID, mode)
	if err != nil {
		h.logger.Error("failed to update notification settings", "userID", userID, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"mode": string(updated)})
}
