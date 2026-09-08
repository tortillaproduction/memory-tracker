package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

var ErrSessionNotFound = errors.New("session not found")

const SessionTTL = 30 * 24 * time.Hour // 30 days

// SessionStore はCookieに入れるセッションIDと、サーバー側で保持するセッション実体を扱う。
// JWTと違いサーバー側で即時失効させられるのが利点。
type SessionStore struct {
	db *sql.DB
}

func NewSessionStore(db *sql.DB) *SessionStore {
	return &SessionStore{db: db}
}

// Create は新しいセッションを発行し、Cookieに入れるべきセッションIDを返す。
func (s *SessionStore) Create(ctx context.Context, userID user.ID) (sessionID string, err error) {
	sessionID, err = generateSessionID()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(SessionTTL)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO sessions (id, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, sessionID, userID, expiresAt)
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

// FindUserID はセッションIDからユーザーIDを引く。期限切れの場合はErrSessionNotFoundを返す。
func (s *SessionStore) FindUserID(ctx context.Context, sessionID string) (user.ID, error) {
	var userID string
	var expiresAt time.Time
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id, expires_at
		FROM sessions
		WHERE id = $1
	`, sessionID).Scan(&userID, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrSessionNotFound
	}
	if err != nil {
		return "", err
	}

	if time.Now().After(expiresAt) {
		return "", ErrSessionNotFound
	}

	return user.ID(userID), nil
}

// Delete はログアウト時にセッションを失効させる。
func (s *SessionStore) Delete(ctx context.Context, sessionID string) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM sessions WHERE id = $1
	`, sessionID)
	return err
}

func generateSessionID() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
