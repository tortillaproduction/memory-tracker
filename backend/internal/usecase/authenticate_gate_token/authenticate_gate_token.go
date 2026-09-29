package authenticate_gate_token

import (
	"context"
	"errors"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/gate"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

var ErrInvalidToken = errors.New("invalid gate token")

// touchInterval より前にlast_used_atを更新していれば、今回は更新しない。
// ゲートは対象アプリを開くたびにAPIを呼ぶため、毎回の書き込みを避ける。
// 通知バッチの「端末未接続」判定(数日単位)には十分な粒度。
const touchInterval = time.Hour

type Usecase struct {
	tokenRepo gate.TokenRepository
	hash      func(plain string) string
}

func NewUsecase(tokenRepo gate.TokenRepository, hash func(plain string) string) *Usecase {
	return &Usecase{tokenRepo: tokenRepo, hash: hash}
}

// Execute は平文トークンを検証し、持ち主のユーザーIDを返す。
func (uc *Usecase) Execute(ctx context.Context, plain string, now time.Time) (user.ID, error) {
	if plain == "" {
		return "", ErrInvalidToken
	}
	t, err := uc.tokenRepo.FindByHash(ctx, uc.hash(plain))
	if errors.Is(err, gate.ErrTokenNotFound) {
		return "", ErrInvalidToken
	}
	if err != nil {
		return "", err
	}

	if last := t.LastUsedAt(); last == nil || now.Sub(*last) >= touchInterval {
		if err := uc.tokenRepo.TouchLastUsed(ctx, t.ID(), now); err != nil {
			return "", err
		}
	}
	return t.UserID(), nil
}
