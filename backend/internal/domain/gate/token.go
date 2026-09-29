package gate

import (
	"context"
	"errors"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

var ErrTokenNotFound = errors.New("gate token not found")

type TokenID string

// Token はAndroidアプリ(ゲート)がAPIを呼ぶための認証トークン。
// 平文はユーザーに一度だけ見せ、ここにはハッシュだけを持つ。
type Token struct {
	id         TokenID
	userID     user.ID
	tokenHash  string
	createdAt  time.Time
	lastUsedAt *time.Time
}

func NewToken(id TokenID, userID user.ID, tokenHash string, createdAt time.Time) *Token {
	return &Token{id: id, userID: userID, tokenHash: tokenHash, createdAt: createdAt}
}

// ReconstructToken はDBから読み出した値でTokenを復元する。
func ReconstructToken(id TokenID, userID user.ID, tokenHash string, createdAt time.Time, lastUsedAt *time.Time) *Token {
	return &Token{id: id, userID: userID, tokenHash: tokenHash, createdAt: createdAt, lastUsedAt: lastUsedAt}
}

func (t *Token) ID() TokenID            { return t.id }
func (t *Token) UserID() user.ID        { return t.userID }
func (t *Token) TokenHash() string      { return t.tokenHash }
func (t *Token) CreatedAt() time.Time   { return t.createdAt }
func (t *Token) LastUsedAt() *time.Time { return t.lastUsedAt }

// TokenRepository はゲート用トークンの永続化を担う。
type TokenRepository interface {
	// Replace はユーザーの既存トークンを新しいトークンで置き換える(無ければ作成する)。
	Replace(ctx context.Context, t *Token) error
	FindByHash(ctx context.Context, tokenHash string) (*Token, error)
	FindByUserID(ctx context.Context, userID user.ID) (*Token, error)
	TouchLastUsed(ctx context.Context, id TokenID, at time.Time) error
}
