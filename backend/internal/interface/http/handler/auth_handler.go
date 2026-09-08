package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
	"github.com/tortillaproduction/memory-tracker/internal/infrastructure/auth"
	"github.com/tortillaproduction/memory-tracker/internal/interface/http/middleware"
	usecaseauth "github.com/tortillaproduction/memory-tracker/internal/usecase/auth"
)

const oauthStateCookieName = "oauth_state"

type AuthHandler struct {
	googleClient  *auth.GoogleOAuthClient
	sessionStore  *auth.SessionStore
	googleLoginUC *usecaseauth.Usecase
	userRepo      user.Repository
	frontendURL   string
	logger        *slog.Logger
}

func NewAuthHandler(
	googleClient *auth.GoogleOAuthClient,
	sessionStore *auth.SessionStore,
	googleLoginUC *usecaseauth.Usecase,
	userRepo user.Repository,
	frontendURL string,
	logger *slog.Logger,
) *AuthHandler {
	return &AuthHandler{
		googleClient:  googleClient,
		sessionStore:  sessionStore,
		googleLoginUC: googleLoginUC,
		userRepo:      userRepo,
		frontendURL:   frontendURL,
		logger:        logger,
	}
}

// LoginRedirect はフロントの「Googleでログイン」ボタンから叩かれる。
// CSRF対策のstateを生成し、短命なCookieに保存した上でGoogleの同意画面へリダイレクトする。
func (h *AuthHandler) LoginRedirect(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // 本番環境ではtrueにする（HTTPS前提）
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((5 * time.Minute).Seconds()),
	})

	http.Redirect(w, r, h.googleClient.AuthURL(state), http.StatusFound)
}

// Callback はGoogleからのリダイレクトを受け、認可コードをユーザー情報に交換した上で
// アプリ独自のセッションを発行し、フロントエンドへリダイレクトする。
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie(oauthStateCookieName)
	if err != nil || r.URL.Query().Get("state") != stateCookie.Value {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	googleUser, err := h.googleClient.ExchangeCode(r.Context(), code)
	if err != nil {
		h.logger.Error("failed to exchange google code", "error", err)
		http.Error(w, "authentication failed", http.StatusBadGateway)
		return
	}

	u, err := h.googleLoginUC.Execute(r.Context(), usecaseauth.GoogleUserInfo{
		GoogleID: googleUser.ID,
		Email:    googleUser.Email,
		Name:     googleUser.Name,
	})
	if err != nil {
		h.logger.Error("failed to find or create user", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	sessionID, err := h.sessionStore.Create(r.Context(), u.ID())
	if err != nil {
		h.logger.Error("failed to create session", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SeesionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // 本番環境ではtrueにする
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(auth.SessionTTL.Seconds()),
	})

	http.Redirect(w, r, h.frontendURL, http.StatusFound)
}

// Me はフロントエンドが起動時に叩き、ログイン状態とユーザー情報を確認するためのAPI。
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	u, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":       u.ID(),
		"email":    u.Email(),
		"name":     u.Name(),
		"planType": u.Plan().Type(),
		"maxSites": u.Plan().MaxSites(),
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(middleware.SeesionCookieName); err == nil {
		_ = h.sessionStore.Delete(r.Context(), cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SeesionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusNoContent)
}

func generateState() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
