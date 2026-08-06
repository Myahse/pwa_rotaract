package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotaract-civ/backend/internal/domain"
)

type AccessRequestRepository struct {
	pool *pgxpool.Pool
}

func NewAccessRequestRepository(pool *pgxpool.Pool) *AccessRequestRepository {
	return &AccessRequestRepository{pool: pool}
}

func (r *AccessRequestRepository) Create(ctx context.Context, req *domain.AccessRequest, passwordHash string) error {
	query := `
		INSERT INTO access_requests (
			club_id, club_name, requested_role_id, email, password_hash,
			first_name, last_name, phone, birth_date, profession, member_since
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, status, created_at
	`
	err := r.pool.QueryRow(ctx, query,
		req.ClubID, req.ClubName, req.RequestedRoleID, req.Email, passwordHash,
		req.FirstName, req.LastName, req.Phone, req.BirthDate, req.Profession, req.MemberSince,
	).Scan(&req.ID, &req.Status, &req.CreatedAt)
	if err != nil {
		return fmt.Errorf("create access request: %w", err)
	}
	return nil
}

func (r *AccessRequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.AccessRequest, error) {
	return r.scanRequest(r.pool.QueryRow(ctx, `
		SELECT id, club_id, club_name, requested_role_id, email,
		       first_name, last_name, phone, birth_date, profession, member_since,
		       status, reviewed_by, review_note, reviewed_at, created_at
		FROM access_requests WHERE id = $1
	`, id))
}

func (r *AccessRequestRepository) List(ctx context.Context, clubID *uuid.UUID, status *domain.AccessRequestStatus) ([]domain.AccessRequest, error) {
	query := `
		SELECT id, club_id, club_name, requested_role_id, email,
		       first_name, last_name, phone, birth_date, profession, member_since,
		       status, reviewed_by, review_note, reviewed_at, created_at
		FROM access_requests
		WHERE 1=1
	`
	args := make([]any, 0, 2)
	if clubID != nil {
		args = append(args, *clubID)
		query += fmt.Sprintf(" AND club_id = $%d", len(args))
	}
	if status != nil {
		args = append(args, *status)
		query += fmt.Sprintf(" AND status = $%d", len(args))
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list access requests: %w", err)
	}
	defer rows.Close()

	requests := make([]domain.AccessRequest, 0)
	for rows.Next() {
		item, err := r.scanRequestRow(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, *item)
	}
	return requests, rows.Err()
}

func (r *AccessRequestRepository) HasPendingByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM access_requests
			WHERE LOWER(email) = LOWER($1) AND status = 'pending'
		)
	`, strings.TrimSpace(email)).Scan(&exists)
	return exists, err
}

func (r *AccessRequestRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.AccessRequestStatus, reviewerID uuid.UUID, note *string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE access_requests
		SET status = $2, reviewed_by = $3, review_note = $4, reviewed_at = NOW()
		WHERE id = $1 AND status = 'pending'
	`, id, status, reviewerID, note)
	if err != nil {
		return fmt.Errorf("update access request: %w", err)
	}
	return nil
}

func (r *AccessRequestRepository) GetPasswordHash(ctx context.Context, id uuid.UUID) (string, error) {
	var hash string
	err := r.pool.QueryRow(ctx, `SELECT password_hash FROM access_requests WHERE id = $1`, id).Scan(&hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return hash, err
}

func (r *AccessRequestRepository) scanRequest(row pgx.Row) (*domain.AccessRequest, error) {
	var req domain.AccessRequest
	err := row.Scan(
		&req.ID, &req.ClubID, &req.ClubName, &req.RequestedRoleID, &req.Email,
		&req.FirstName, &req.LastName, &req.Phone, &req.BirthDate, &req.Profession, &req.MemberSince,
		&req.Status, &req.ReviewedBy, &req.ReviewNote, &req.ReviewedAt, &req.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan access request: %w", err)
	}
	return &req, nil
}

func (r *AccessRequestRepository) scanRequestRow(rows pgx.Rows) (*domain.AccessRequest, error) {
	var req domain.AccessRequest
	err := rows.Scan(
		&req.ID, &req.ClubID, &req.ClubName, &req.RequestedRoleID, &req.Email,
		&req.FirstName, &req.LastName, &req.Phone, &req.BirthDate, &req.Profession, &req.MemberSince,
		&req.Status, &req.ReviewedBy, &req.ReviewNote, &req.ReviewedAt, &req.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan access request: %w", err)
	}
	return &req, nil
}
