package issue_gate_token

import (
	"context"
	"errors"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/gate"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

type TokenGenerator interface {
	Generate() (plain, hash string, err error)
}

type IDGenerator interface {
	NewGateTokenID() gate.TokenID
}

// Status はWebの設定画面に表示するトークンの状態。平文は含まない。
type Status struct {
	Exists     bool
	CreatedAt  *time.Time
	LastUsedAt *time.Time
}

type Usecase struct {
	tokenRepo   gate.TokenRepository
	generator   TokenGenerator
	idGenerator IDGenerator
}

func NewUsecase(tokenRepo gate.TokenRepository, generator TokenGenerator, idGen IDGenerator) *Usecase {
	return &Usecase{tokenRepo: tokenRepo, generator: generator, idGenerator: idGen}
}

// Execute はトークンを発行(既にあれば再発行)し、平文を返す。
// 再発行すると以前のトークンは使えなくなる。平文はここでしか取得できない。
func (uc *Usecase) Execute(ctx context.Context, userID user.ID, now time.Time) (string, error) {
	plain, hash, err := uc.generator.Generate()
	if err != nil {
		return "", err
	}
	t := gate.NewToken(uc.idGenerator.NewGateTokenID(), userID, hash, now)
	if err := uc.tokenRepo.Replace(ctx, t); err != nil {
		return "", err
	}
	return plain, nil
}

func (uc *Usecase) Status(ctx context.Context, userID user.ID) (Status, error) {
	t, err := uc.tokenRepo.FindByUserID(ctx, userID)
	if errors.Is(err, gate.ErrTokenNotFound) {
		return Status{}, nil
	}
	if err != nil {
		return Status{}, err
	}
	createdAt := t.CreatedAt()
	return Status{Exists: true, CreatedAt: &createdAt, LastUsedAt: t.LastUsedAt()}, nil
}
