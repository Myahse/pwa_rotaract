package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotaract-civ/backend/internal/domain"
)

type ClubRegistrationRepository struct {
	pool *pgxpool.Pool
}

func NewClubRegistrationRepository(pool *pgxpool.Pool) *ClubRegistrationRepository {
	return &ClubRegistrationRepository{pool: pool}
}

const clubRegistrationSelectColumns = `
	id, club_name, contact_email, contact_first_name, contact_last_name,
	phone, description, country, city, commune, founded_at, message,
	status, reviewed_by, review_note, reviewed_at, created_club_id,
	access_token, access_token_expires_at, access_token_used_at, approved_slug, created_at
`

func (r *ClubRegistrationRepository) Create(ctx context.Context, req *domain.ClubRegistrationRequest) error {
	query := `
		INSERT INTO club_registration_requests (
			club_name, contact_email, contact_first_name, contact_last_name,
			phone, description, country, city, commune, founded_at, message
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, status, created_at
	`
	err := r.pool.QueryRow(ctx, query,
		req.ClubName, req.ContactEmail, req.ContactFirstName, req.ContactLastName,
		req.Phone, req.Description, req.Country, req.City, req.Commune, req.FoundedAt, req.Message,
	).Scan(&req.ID, &req.Status, &req.CreatedAt)
	if err != nil {
		return fmt.Errorf("create club registration request: %w", err)
	}
	return nil
}

func (r *ClubRegistrationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ClubRegistrationRequest, error) {
	return r.scanRequest(r.pool.QueryRow(ctx, `
		SELECT `+clubRegistrationSelectColumns+`
		FROM club_registration_requests
		WHERE id = $1
	`, id))
}

func (r *ClubRegistrationRepository) GetByAccessToken(ctx context.Context, token string) (*domain.ClubRegistrationRequest, error) {
	return r.scanRequest(r.pool.QueryRow(ctx, `
		SELECT `+clubRegistrationSelectColumns+`
		FROM club_registration_requests
		WHERE access_token = $1
	`, strings.TrimSpace(token)))
}

func (r *ClubRegistrationRepository) List(ctx context.Context, status *domain.ClubRegistrationStatus) ([]domain.ClubRegistrationRequest, error) {
	query := `
		SELECT ` + clubRegistrationSelectColumns + `
		FROM club_registration_requests
		WHERE 1=1
	`
	args := make([]any, 0, 1)
	if status != nil {
		args = append(args, *status)
		query += fmt.Sprintf(" AND status = $%d", len(args))
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list club registration requests: %w", err)
	}
	defer rows.Close()

	requests := make([]domain.ClubRegistrationRequest, 0)
	for rows.Next() {
		item, err := r.scanRequestRow(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, *item)
	}
	return requests, rows.Err()
}

func (r *ClubRegistrationRepository) HasPendingByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM club_registration_requests
			WHERE LOWER(contact_email) = LOWER($1) AND status = 'pending'
		)
	`, strings.TrimSpace(email)).Scan(&exists)
	return exists, err
}

func (r *ClubRegistrationRepository) HasPendingByClubName(ctx context.Context, clubName string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM club_registration_requests
			WHERE LOWER(club_name) = LOWER($1) AND status = 'pending'
		)
	`, strings.TrimSpace(clubName)).Scan(&exists)
	return exists, err
}

type ApproveRegistrationParams struct {
	ReviewerID uuid.UUID
	Note       *string
	Token      string
	ExpiresAt  time.Time
	Slug       *string
	Country    string
	City       string
	Commune    string
}

func (r *ClubRegistrationRepository) MarkApproved(ctx context.Context, id uuid.UUID, params ApproveRegistrationParams) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE club_registration_requests
		SET status = 'approved',
		    reviewed_by = $2,
		    review_note = $3,
		    reviewed_at = NOW(),
		    access_token = $4,
		    access_token_expires_at = $5,
		    access_token_used_at = NULL,
		    approved_slug = $6,
		    country = $7,
		    city = $8,
		    commune = $9
		WHERE id = $1 AND status = 'pending'
	`, id, params.ReviewerID, params.Note, params.Token, params.ExpiresAt, params.Slug,
		params.Country, params.City, params.Commune)
	if err != nil {
		return fmt.Errorf("approve club registration request: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ClubRegistrationRepository) MarkRejected(ctx context.Context, id, reviewerID uuid.UUID, note *string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE club_registration_requests
		SET status = 'rejected',
		    reviewed_by = $2,
		    review_note = $3,
		    reviewed_at = NOW()
		WHERE id = $1 AND status = 'pending'
	`, id, reviewerID, note)
	if err != nil {
		return fmt.Errorf("reject club registration request: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ClubRegistrationRepository) MarkCompleted(ctx context.Context, id, clubID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE club_registration_requests
		SET created_club_id = $2,
		    access_token_used_at = NOW()
		WHERE id = $1
		  AND status = 'approved'
		  AND access_token_used_at IS NULL
		  AND created_club_id IS NULL
	`, id, clubID)
	if err != nil {
		return fmt.Errorf("complete club registration request: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func GenerateAccessToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (r *ClubRegistrationRepository) scanRequest(row pgx.Row) (*domain.ClubRegistrationRequest, error) {
	var req domain.ClubRegistrationRequest
	var accessToken *string
	err := row.Scan(
		&req.ID, &req.ClubName, &req.ContactEmail, &req.ContactFirstName, &req.ContactLastName,
		&req.Phone, &req.Description, &req.Country, &req.City, &req.Commune, &req.FoundedAt, &req.Message,
		&req.Status, &req.ReviewedBy, &req.ReviewNote, &req.ReviewedAt, &req.CreatedClubID,
		&accessToken, &req.AccessTokenExpiresAt, &req.AccessTokenUsedAt, &req.ApprovedSlug, &req.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan club registration request: %w", err)
	}
	if accessToken != nil {
		req.AccessToken = *accessToken
	}
	return &req, nil
}

func (r *ClubRegistrationRepository) scanRequestRow(rows pgx.Rows) (*domain.ClubRegistrationRequest, error) {
	var req domain.ClubRegistrationRequest
	var accessToken *string
	err := rows.Scan(
		&req.ID, &req.ClubName, &req.ContactEmail, &req.ContactFirstName, &req.ContactLastName,
		&req.Phone, &req.Description, &req.Country, &req.City, &req.Commune, &req.FoundedAt, &req.Message,
		&req.Status, &req.ReviewedBy, &req.ReviewNote, &req.ReviewedAt, &req.CreatedClubID,
		&accessToken, &req.AccessTokenExpiresAt, &req.AccessTokenUsedAt, &req.ApprovedSlug, &req.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan club registration request: %w", err)
	}
	if accessToken != nil {
		req.AccessToken = *accessToken
	}
	return &req, nil
}
