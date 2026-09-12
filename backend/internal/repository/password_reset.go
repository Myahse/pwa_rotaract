package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PasswordResetToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
}

type PasswordResetRepository struct{ pool *pgxpool.Pool }

func NewPasswordResetRepository(pool *pgxpool.Pool) *PasswordResetRepository {
	return &PasswordResetRepository{pool: pool}
}

func (r *PasswordResetRepository) Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM password_reset_tokens WHERE user_id = $1`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM password_reset_tokens WHERE expires_at < NOW()`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PasswordResetRepository) GetValid(ctx context.Context, tokenHash string) (*PasswordResetToken, error) {
	var token PasswordResetToken
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, used_at
		FROM password_reset_tokens
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()
	`, tokenHash).Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.UsedAt)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *PasswordResetRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1 AND used_at IS NULL`, id)
	return err
}
