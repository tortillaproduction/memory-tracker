package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/tortillaproduction/memory-tracker/internal/domain/gate"
	"github.com/tortillaproduction/memory-tracker/internal/domain/user"
)

type gateTokenRepository struct {
	db *sql.DB
}

func NewGateTokenRepository(db *sql.DB) gate.TokenRepository {
	return &gateTokenRepository{db: db}
}

func (r *gateTokenRepository) Replace(ctx context.Context, t *gate.Token) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO gate_tokens (id, user_id, token_hash, created_at, last_used_at)
		VALUES ($1, $2, $3, $4, NULL)
		ON CONFLICT (user_id) DO UPDATE
		SET id = EXCLUDED.id, token_hash = EXCLUDED.token_hash,
		    created_at = EXCLUDED.created_at, last_used_at = NULL
	`, t.ID(), t.UserID(), t.TokenHash(), t.CreatedAt())
	return err
}

func (r *gateTokenRepository) FindByHash(ctx context.Context, tokenHash string) (*gate.Token, error) {
	return r.findOne(ctx, `WHERE token_hash = $1`, tokenHash)
}

func (r *gateTokenRepository) FindByUserID(ctx context.Context, userID user.ID) (*gate.Token, error) {
	return r.findOne(ctx, `WHERE user_id = $1`, userID)
}

func (r *gateTokenRepository) findOne(ctx context.Context, where string, arg any) (*gate.Token, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, created_at, last_used_at
		FROM gate_tokens `+where, arg)

	var id, userID, tokenHash string
	var createdAt time.Time
	var lastUsedAt sql.NullTime
	if err := row.Scan(&id, &userID, &tokenHash, &createdAt, &lastUsedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, gate.ErrTokenNotFound
		}
		return nil, err
	}

	var lastUsed *time.Time
	if lastUsedAt.Valid {
		lastUsed = &lastUsedAt.Time
	}
	return gate.ReconstructToken(gate.TokenID(id), user.ID(userID), tokenHash, createdAt, lastUsed), nil
}

func (r *gateTokenRepository) TouchLastUsed(ctx context.Context, id gate.TokenID, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE gate_tokens SET last_used_at = $2 WHERE id = $1`, id, at)
	return err
}
