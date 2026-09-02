package site

import (
	"context"

	"github.com/tortillaproduction/study-tracker/internal/domain/user"
)

// Repository はドメイン層で定義するインターフェース。
// 実装(infrastructure/persistence/postgres)はこのインターフェースに依存性逆転する。
type Repository interface {
	Save(ctx context.Context, s *Site) error
	FindByID(ctx context.Context, id ID) (*Site, error)
	FindAllByUserID(ctx context.Context, userID user.ID) ([]*Site, error)
	CountByUserID(ctx context.Context, userID user.ID) (int, error)
	Delete(ctx context.Context, id ID) error
}
