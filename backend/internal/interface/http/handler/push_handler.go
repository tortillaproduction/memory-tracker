package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/tortillaproduction/memory-tracker/internal/usecase/subscribe_push"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/unsubscribe_push"
	"github.com/tortillaproduction/memory-tracker/internal/usecase/update_notification_preferences"
)

type PushHandler struct {
	vapidPublicKey string
	subscribeUC    *subscribe_push.Usecase
	unsubscribeUC  *unsubscribe_push.Usecase
	preferencesUC  *update_notification_preferences.Usecase
	logger         *slog.Logger
}

func NewPushHandler(
	vapidPublicKey string,
	subscribeUC *subscribe_push.Usecase,
	unsubscribeUC *unsubscribe_push.Usecase,
	preferencesUC *update_notification_preferences.Usecase,
	logger *slog.Logger,
) *PushHandler {
	return &PushHandler{
		vapidPublicKey: vapidPublicKey,
		subscribeUC:    subscribeUC,
		unsubscribeUC:  unsubscribeUC,
		preferencesUC:  preferencesUC,
		logger:         logger,
	}
}

// VAPIDPublicKey は GET /api/push/vapid-public-key のハンドラー(認証不要)。
// フロントエンドはこの公開鍵を pushManager.subscribe の applicationServerKey に使う。
func (h *PushHandler) VAPIDPublicKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"publicKey": h.vapidPublicKey})
}

// Subscribe は POST /api/push/subscribe のハンドラー。
// フロントエンドの PushSubscription (endpoint/keys) をそのまま受け取る。
func (h *PushHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Endpoint string `json:"endpoint"`
		Keys     struct {
			P256dh string `json:"p256dh"`
			Auth   string `json:"auth"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Endpoint == "" || req.Keys.P256dh == "" || req.Keys.Auth == "" {
		http.Error(w, "endpoint and keys are required", http.StatusBadRequest)
		return
	}

	userID := userIDFromContext(r.Context())

	err := h.subscribeUC.Execute(r.Context(), subscribe_push.Input{
		UserID:    userID,
		Endpoint:  req.Endpoint,
		P256dhKey: req.Keys.P256dh,
		AuthKey:   req.Keys.Auth,
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		h.logger.Error("failed to subscribe push", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Unsubscribe は POST /api/push/unsubscribe のハンドラー。
func (h *PushHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Endpoint string `json:"endpoint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Endpoint == "" {
		http.Error(w, "endpoint is required", http.StatusBadRequest)
		return
	}

	userID := userIDFromContext(r.Context())

	if err := h.unsubscribeUC.Execute(r.Context(), unsubscribe_push.Input{UserID: userID, Endpoint: req.Endpoint}); err != nil {
		h.logger.Error("failed to unsubscribe push", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type notificationPreferencesResponse struct {
	EmailEnabled                  bool `json:"emailEnabled"`
	PushEnabled                   bool `json:"pushEnabled"`
	DisableEmailWhenPushAvailable bool `json:"disableEmailWhenPushAvailable"`
}

// GetPreferences は GET /api/notification-preferences のハンドラー。
func (h *PushHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	prefs, err := h.preferencesUC.Get(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get notification preferences", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notificationPreferencesResponse{
		EmailEnabled:                  prefs.EmailEnabled,
		PushEnabled:                   prefs.PushEnabled,
		DisableEmailWhenPushAvailable: prefs.DisableEmailWhenPushAvailable,
	})
}

// UpdatePreferences は PATCH /api/notification-preferences のハンドラー。
// 現状は「プッシュが使える場合はメールを止める」設定のみ変更可能。
func (h *PushHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DisableEmailWhenPushAvailable *bool `json:"disableEmailWhenPushAvailable"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.DisableEmailWhenPushAvailable == nil {
		http.Error(w, "disableEmailWhenPushAvailable is required", http.StatusBadRequest)
		return
	}

	userID := userIDFromContext(r.Context())

	prefs, err := h.preferencesUC.UpdateDisableEmailWhenPushAvailable(r.Context(), userID, *req.DisableEmailWhenPushAvailable)
	if err != nil {
		h.logger.Error("failed to update notification preferences", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notificationPreferencesResponse{
		EmailEnabled:                  prefs.EmailEnabled,
		PushEnabled:                   prefs.PushEnabled,
		DisableEmailWhenPushAvailable: prefs.DisableEmailWhenPushAvailable,
	})
}
