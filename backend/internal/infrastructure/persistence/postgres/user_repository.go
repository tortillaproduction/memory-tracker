package postgres

import (
	"context"
	"database/sql"

	"github.com/tortillaproduction/study-tracker/internal/domain/user"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) user.Repository {
	return &userRepository{db: db}
}

func (r *userRepository) Save(ctx context.Context, u *user.User) error {
	// TODO: 実装
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id user.ID) (*user.User, error) {
	// TODO: 実装
	return nil, nil
}

func (r *userRepository) FindByGoogleID(ctx context.Context, googleID string) (*user.User, error) {
	// TODO: 実装
	return nil, nil
}
