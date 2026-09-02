package user

import "context"

type Repository interface {
	Save(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id ID) (*User, error)
	FindByGoogleID(ctx context.Context, googleID string) (*User, error)
}
