package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotaract-civ/backend/internal/domain"
)

type DonationRepository struct {
	pool *pgxpool.Pool
}

func NewDonationRepository(pool *pgxpool.Pool) *DonationRepository {
	return &DonationRepository{pool: pool}
}

func (r *DonationRepository) Create(ctx context.Context, donation *domain.Donation) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO donations (user_id, name, email, amount_xof)
		VALUES ($1, $2, $3, $4)
		RETURNING id, status, created_at
	`, donation.UserID, donation.Name, donation.Email, donation.AmountXOF).
		Scan(&donation.ID, &donation.Status, &donation.CreatedAt)
	if err != nil {
		return fmt.Errorf("create donation: %w", err)
	}
	donation.KnownMember = donation.UserID != nil
	return nil
}

func (r *DonationRepository) SetReceipt(ctx context.Context, id uuid.UUID, path, mime string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE donations SET receipt_path = $2, receipt_mime = $3 WHERE id = $1
	`, id, path, mime)
	if err != nil {
		return fmt.Errorf("set donation receipt: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *DonationRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Donation, error) {
	item, err := scanDonation(r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, email, amount_xof, status, receipt_path, receipt_mime, created_at
		FROM donations WHERE id = $1
	`, id))
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get donation: %w", err)
	}
	return item, nil
}

func (r *DonationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM donations WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete donation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *DonationRepository) List(ctx context.Context) ([]domain.Donation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, email, amount_xof, status, receipt_path, receipt_mime, created_at
		FROM donations
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list donations: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Donation, 0)
	for rows.Next() {
		item, err := scanDonation(rows)
		if err != nil {
			return nil, fmt.Errorf("scan donation: %w", err)
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *DonationRepository) MarkReceived(ctx context.Context, id uuid.UUID) (*domain.Donation, error) {
	item, err := scanDonation(r.pool.QueryRow(ctx, `
		UPDATE donations
		SET status = 'received'
		WHERE id = $1 AND status = 'pending'
		RETURNING id, user_id, name, email, amount_xof, status, receipt_path, receipt_mime, created_at
	`, id))
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("mark donation received: %w", err)
	}
	return item, nil
}

func scanDonation(row pgx.Row) (*domain.Donation, error) {
	var item domain.Donation
	if err := row.Scan(
		&item.ID, &item.UserID, &item.Name, &item.Email, &item.AmountXOF,
		&item.Status, &item.ReceiptPath, &item.ReceiptMime, &item.CreatedAt,
	); err != nil {
		return nil, err
	}
	item.KnownMember = item.UserID != nil
	return &item, nil
}
