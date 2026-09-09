package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/tortillaproduction/memory-tracker/internal/domain/plan"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) user.Repository {
	return &userRepository{db: db}
}

func (r *userRepository) Save(ctx context.Context, u *user.User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, google_id, email, name, picture_url, plan_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
		  email = EXCLUDED.email,
		  name = EXCLUDED.name,
		  picture_url = EXCLUDED.picture_url,
		  plan_type = EXCLUDED.plan_type
	`, u.ID(), u.GoogleID(), u.Email(), u.Name(), u.PictureURL(), string(u.Plan().Type()))
	return err
}

func (r *userRepository) FindByID(ctx context.Context, id user.ID) (*user.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, google_id, email, name, picture_url, plan_type
		FROM users WHERE id = $1
	`, id)

	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, user.ErrNotFound
	}
	return u, err
}

func (r *userRepository) FindByGoogleID(ctx context.Context, googleID string) (*user.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, google_id, email, name, picture_url, plan_type
		FROM users WHERE google_id = $1
	`, googleID)

	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, user.ErrNotFound
	}
	return u, err
}

func scanUser(row *sql.Row) (*user.User, error) {
	var id, googleID, email, name, planType string
	var pictureURL sql.NullString
	if err := row.Scan(&id, &googleID, &email, &name, &pictureURL, &planType); err != nil {
		return nil, err
	}

	u := user.NewUser(user.ID(id), googleID, email, name, pictureURL.String)
	if planType == string(plan.TypePremium) {
		u.UpgradeTo(plan.Premium())
	}

	return u, nil
}
