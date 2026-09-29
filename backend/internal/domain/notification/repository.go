package notification

import (
	"context"
	"errors"

	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

var ErrSettingNotFound = errors.New("notification setting not found")

type SettingRepository interface {
	Create(ctx context.Context, s *Setting) error
	FindByUserID(ctx context.Context, userID user.ID) (*Setting, error)
	Update(ctx context.Context, s *Setting) error
}
