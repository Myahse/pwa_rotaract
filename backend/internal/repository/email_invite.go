package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotaract-civ/backend/internal/domain"
)

type EmailInviteRepository struct {
	pool *pgxpool.Pool
}

func NewEmailInviteRepository(pool *pgxpool.Pool) *EmailInviteRepository {
	return &EmailInviteRepository{pool: pool}
}

func (r *EmailInviteRepository) Create(ctx context.Context, invite *domain.EmailInvite) error {
	token, err := generateToken(32)
	if err != nil {
		return err
	}
	invite.Token = token

	query := `
		INSERT INTO email_invites (club_id, email, role_id, token, invited_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	err = r.pool.QueryRow(ctx, query,
		invite.ClubID, strings.ToLower(strings.TrimSpace(invite.Email)), invite.RoleID,
		invite.Token, invite.InvitedBy, invite.ExpiresAt,
	).Scan(&invite.ID, &invite.CreatedAt)
	if err != nil {
		return fmt.Errorf("create email invite: %w", err)
	}
	return nil
}

func (r *EmailInviteRepository) GetByToken(ctx context.Context, token string) (*domain.EmailInvite, error) {
	query := `
		SELECT id, club_id, email, role_id, token, invited_by, expires_at, used_at, created_at
		FROM email_invites WHERE token = $1
	`
	return r.scanInvite(r.pool.QueryRow(ctx, query, strings.TrimSpace(token)))
}

func (r *EmailInviteRepository) HasPending(ctx context.Context, clubID uuid.UUID, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM email_invites
			WHERE club_id = $1 AND LOWER(email) = LOWER($2)
			  AND used_at IS NULL AND expires_at > NOW()
		)
	`, clubID, strings.TrimSpace(email)).Scan(&exists)
	return exists, err
}

func (r *EmailInviteRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE email_invites SET used_at = NOW()
		WHERE id = $1 AND used_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("mark invite used: %w", err)
	}
	return nil
}

func (r *EmailInviteRepository) ListByClub(ctx context.Context, clubID uuid.UUID) ([]domain.EmailInvite, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, club_id, email, role_id, token, invited_by, expires_at, used_at, created_at
		FROM email_invites
		WHERE club_id = $1
		ORDER BY created_at DESC
	`, clubID)
	if err != nil {
		return nil, fmt.Errorf("list email invites: %w", err)
	}
	defer rows.Close()

	invites := make([]domain.EmailInvite, 0)
	for rows.Next() {
		item, err := r.scanInviteRow(rows)
		if err != nil {
			return nil, err
		}
		invites = append(invites, *item)
	}
	return invites, rows.Err()
}

func (r *EmailInviteRepository) GetByID(ctx context.Context, clubID, inviteID uuid.UUID) (*domain.EmailInvite, error) {
	query := `
		SELECT id, club_id, email, role_id, token, invited_by, expires_at, used_at, created_at
		FROM email_invites WHERE id = $1 AND club_id = $2
	`
	return r.scanInvite(r.pool.QueryRow(ctx, query, inviteID, clubID))
}

func (r *EmailInviteRepository) RevokePending(ctx context.Context, clubID, inviteID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM email_invites
		WHERE id = $1 AND club_id = $2 AND used_at IS NULL
	`, inviteID, clubID)
	if err != nil {
		return fmt.Errorf("revoke email invite: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *EmailInviteRepository) scanInvite(row pgx.Row) (*domain.EmailInvite, error) {
	var invite domain.EmailInvite
	err := row.Scan(
		&invite.ID, &invite.ClubID, &invite.Email, &invite.RoleID, &invite.Token,
		&invite.InvitedBy, &invite.ExpiresAt, &invite.UsedAt, &invite.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan email invite: %w", err)
	}
	return &invite, nil
}

func (r *EmailInviteRepository) scanInviteRow(rows pgx.Rows) (*domain.EmailInvite, error) {
	var invite domain.EmailInvite
	err := rows.Scan(
		&invite.ID, &invite.ClubID, &invite.Email, &invite.RoleID, &invite.Token,
		&invite.InvitedBy, &invite.ExpiresAt, &invite.UsedAt, &invite.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan email invite: %w", err)
	}
	return &invite, nil
}

func generateToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
