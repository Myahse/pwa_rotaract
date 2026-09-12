package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotaract-civ/backend/internal/domain"
)

type ClubDueRepository struct {
	pool *pgxpool.Pool
}

func NewClubDueRepository(pool *pgxpool.Pool) *ClubDueRepository {
	return &ClubDueRepository{pool: pool}
}

func (r *ClubDueRepository) Create(ctx context.Context, due *domain.ClubDuePayment) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO club_due_payments (club_id, user_id, due_month, amount_xof)
		VALUES ($1, $2, $3::date, $4)
		RETURNING id, status, created_at
	`, due.ClubID, due.UserID, due.DueMonth, due.AmountXOF).
		Scan(&due.ID, &due.Status, &due.CreatedAt)
	if err != nil {
		return fmt.Errorf("create club due: %w", err)
	}
	return nil
}

func (r *ClubDueRepository) SetReceipt(ctx context.Context, id uuid.UUID, path, mime string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE club_due_payments SET receipt_path = $2, receipt_mime = $3 WHERE id = $1
	`, id, path, mime)
	if err != nil {
		return fmt.Errorf("set club due receipt: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ClubDueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM club_due_payments WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ClubDueRepository) Get(ctx context.Context, id uuid.UUID) (*domain.ClubDuePayment, error) {
	return scanClubDue(r.pool.QueryRow(ctx, `
		SELECT d.id, d.club_id, d.user_id, to_char(d.due_month, 'YYYY-MM'), d.amount_xof, d.status,
		       d.receipt_path, d.receipt_mime, d.reviewed_by, d.reviewed_at, d.created_at,
		       u.id, u.email, u.first_name, u.last_name, u.phone, u.birth_date, u.profession,
		       u.member_since, u.avatar_path, u.is_admin, u.is_active, u.created_at, u.updated_at
		FROM club_due_payments d
		JOIN users u ON u.id = d.user_id
		WHERE d.id = $1
	`, id))
}

func (r *ClubDueRepository) ListByClub(ctx context.Context, clubID uuid.UUID) ([]domain.ClubDuePayment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT d.id, d.club_id, d.user_id, to_char(d.due_month, 'YYYY-MM'), d.amount_xof, d.status,
		       d.receipt_path, d.receipt_mime, d.reviewed_by, d.reviewed_at, d.created_at,
		       u.id, u.email, u.first_name, u.last_name, u.phone, u.birth_date, u.profession,
		       u.member_since, u.avatar_path, u.is_admin, u.is_active, u.created_at, u.updated_at
		FROM club_due_payments d
		JOIN users u ON u.id = d.user_id
		WHERE d.club_id = $1
		ORDER BY d.due_month DESC, d.created_at DESC
	`, clubID)
	if err != nil {
		return nil, fmt.Errorf("list club dues: %w", err)
	}
	defer rows.Close()
	return scanClubDues(rows)
}

func (r *ClubDueRepository) ListByClubAndUser(ctx context.Context, clubID, userID uuid.UUID) ([]domain.ClubDuePayment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT d.id, d.club_id, d.user_id, to_char(d.due_month, 'YYYY-MM'), d.amount_xof, d.status,
		       d.receipt_path, d.receipt_mime, d.reviewed_by, d.reviewed_at, d.created_at,
		       u.id, u.email, u.first_name, u.last_name, u.phone, u.birth_date, u.profession,
		       u.member_since, u.avatar_path, u.is_admin, u.is_active, u.created_at, u.updated_at
		FROM club_due_payments d
		JOIN users u ON u.id = d.user_id
		WHERE d.club_id = $1 AND d.user_id = $2
		ORDER BY d.due_month DESC, d.created_at DESC
	`, clubID, userID)
	if err != nil {
		return nil, fmt.Errorf("list user club dues: %w", err)
	}
	defer rows.Close()
	return scanClubDues(rows)
}

func (r *ClubDueRepository) MarkReceived(ctx context.Context, id, reviewerID uuid.UUID) (*domain.ClubDuePayment, error) {
	err := r.pool.QueryRow(ctx, `
		UPDATE club_due_payments
		SET status = 'received', reviewed_by = $2, reviewed_at = NOW()
		WHERE id = $1 AND status = 'pending'
		RETURNING id
	`, id, reviewerID).Scan(&id)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

func scanClubDue(row pgx.Row) (*domain.ClubDuePayment, error) {
	var due domain.ClubDuePayment
	var user domain.User
	var reviewedAt *time.Time
	err := row.Scan(
		&due.ID, &due.ClubID, &due.UserID, &due.DueMonth, &due.AmountXOF, &due.Status,
		&due.ReceiptPath, &due.ReceiptMime, &due.ReviewedBy, &reviewedAt, &due.CreatedAt,
		&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Phone, &user.BirthDate, &user.Profession,
		&user.MemberSince, &user.AvatarPath, &user.IsAdmin, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	due.ReviewedAt = reviewedAt
	due.User = &user
	return &due, nil
}

func scanClubDues(rows pgx.Rows) ([]domain.ClubDuePayment, error) {
	items := make([]domain.ClubDuePayment, 0)
	for rows.Next() {
		var due domain.ClubDuePayment
		var user domain.User
		var reviewedAt *time.Time
		if err := rows.Scan(
			&due.ID, &due.ClubID, &due.UserID, &due.DueMonth, &due.AmountXOF, &due.Status,
			&due.ReceiptPath, &due.ReceiptMime, &due.ReviewedBy, &reviewedAt, &due.CreatedAt,
			&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Phone, &user.BirthDate, &user.Profession,
			&user.MemberSince, &user.AvatarPath, &user.IsAdmin, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		due.ReviewedAt = reviewedAt
		due.User = &user
		items = append(items, due)
	}
	return items, rows.Err()
}
