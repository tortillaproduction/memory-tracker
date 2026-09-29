package dismiss_gate

import (
	"context"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/gate"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

type IDGenerator interface {
	NewGateDismissalID() gate.DismissalID
}

type Usecase struct {
	dismissalRepo gate.DismissalRepository
	idGenerator   IDGenerator
}

func NewUsecase(dismissalRepo gate.DismissalRepository, idGen IDGenerator) *Usecase {
	return &Usecase{dismissalRepo: dismissalRepo, idGenerator: idGen}
}

// Execute は今日(日本時間)ゲートを脱出口で解除したことを記録する。
// 同じ日に何度呼んでも記録は1件で、その日はゲートが再び出なくなる。
func (uc *Usecase) Execute(ctx context.Context, userID user.ID, now time.Time) error {
	return uc.dismissalRepo.Record(ctx, uc.idGenerator.NewGateDismissalID(), userID, gate.StartOfDay(now))
}
