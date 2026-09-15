package notification

import "context"

type Repository interface {
	Create(ctx context.Context, s *Setting) error
}
