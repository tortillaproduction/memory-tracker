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

// Save はgoogle_idをUPSERTのキーにする。IDはリクエストのたびに新規生成される
// ULIDのため主キー競合では検知できず、同一Googleアカウントに対する同時ログイン
// （TOCTOU競合）はgoogle_idのUNIQUE制約でしか捕捉できないため。
// RETURNINGの(xmax = 0)は、このステートメント自身がINSERTを行ったか
// （falseなら既存行の更新＝他方の同時リクエストが先に作成済み）を表す。
func (r *userRepository) Save(ctx context.Context, u *user.User) (*user.User, bool, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO users (id, google_id, email, name, picture_url, plan_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (google_id) DO UPDATE SET
		  email = EXCLUDED.email,
		  name = EXCLUDED.name,
		  picture_url = EXCLUDED.picture_url,
		  plan_type = EXCLUDED.plan_type
		RETURNING id, google_id, email, name, picture_url, plan_type, (xmax = 0) AS inserted
	`, u.ID(), u.GoogleID(), u.Email(), u.Name(), u.PictureURL(), string(u.Plan().Type()))

	var id, googleID, email, name, planType string
	var pictureURL sql.NullString
	var inserted bool
	if err := row.Scan(&id, &googleID, &email, &name, &pictureURL, &planType, &inserted); err != nil {
		return nil, false, err
	}

	return buildUser(id, googleID, email, name, pictureURL, planType), inserted, nil
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

	return buildUser(id, googleID, email, name, pictureURL, planType), nil
}

func buildUser(id, googleID, email, name string, pictureURL sql.NullString, planType string) *user.User {
	u := user.NewUser(user.ID(id), googleID, email, name, pictureURL.String)
	if planType == string(plan.TypePremium) {
		u.UpgradeTo(plan.Premium())
	}
	return u
}
